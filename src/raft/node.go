package raft

import (
	"fmt"
	"math/rand"
	"time"
)

const CurrentTermKey = "current-term"
const timeBetweenHeartbeats = 50

type node struct {
	st              *state
	electionTimeout int64
	id              string

	dispatcher          Dispatcher
	store               Store
	electionExpiryTimer Timer
	time                Time
	campaigner          Campaigner
	whoArePeers         WhoArePeers
	transport           Transport
	votedFor            VotedFor
	log                 Log
	heartbeatTimer      Timer
}

func newNode(id string) *node {
	rand.Seed(time.Now().Unix())
	n := new(node)
	n.id = id
	return n
}

func (n *node) boot() {
	n.st = new(state)
	n.store.StoreInt(CurrentTermKey, 0)
	n.st.mode = Follower
	n.st.commitIndex = 0
	n.st.lastApplied = 0

	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) handleEvent(event event) {
	switch event.eventType {
	case StartFollower:
		n.startFollower(event)
	case ElectionTimerTimedout:
		n.electionTimerTimeout(event)
	case StartCandidate:
		n.startCandidate(event)
	case GotVoteResponse:
		n.gotVote(event)
	case StepDown:
		n.stepDown(event)
	case GotRequestForVote:
		n.gotRequestForVote(event)
	case AppendEntries:
		n.appendEntries(event)
	case StartLeader:
		n.startLeader(event)
	case HeartbeatTimerTimedout:
		n.heartbeatTimerTimedout(event)
	case GotAppendEntriesResponse:
		n.gotAppendEntriesResponse(event)
	default:
		panic(fmt.Sprintf("Unknown event: %d passed to handleEvent", event.eventType))
	}
}

func (n *node) startElectionTimer() {
	n.electionTimeout = getRandomizedElectionTimout()
	n.electionExpiryTimer.Start(time.Duration(n.electionTimeout) * time.Millisecond)
}

func (n *node) restartElectionTimer() {
	n.electionExpiryTimer.Stop()
	n.startElectionTimer()
}

func (n *node) startFollower(evt event) {
	if n.st.mode != Follower {
		panic("Mode is not set to follower in startFollower")
	}
	n.startElectionTimer()
}

func (n *node) hasHeardFromALeader() bool {
	return (n.time.UnixNow() - n.st.lastHeardFromALeader) < n.electionTimeout
}

func (n *node) electionTimerTimeout(evt event) {
	if n.hasHeardFromALeader() {
		return
	}
	if n.st.mode == Follower {
		n.st.mode = Candidate
		n.dispatcher.Dispatch(event{StartCandidate, nil})
	} else if n.st.mode == Candidate {
		// think about back off from: ARC Heidi Howard - OCAML raft
		n.st.mode = Candidate
		n.dispatcher.Dispatch(event{StartCandidate, nil})
	} else {
		panic("Received election timer timedout while being a leader")
	}
}

func (n *node) startCandidate(event event) {
	if n.st.mode != Candidate {
		panic("Mode is not set to candidate in startCandidate")
	}
	term := getCurrentTerm(n)
	term = term + 1
	n.store.StoreInt(CurrentTermKey, term)
	n.st.votesGot = n.st.votesGot + 1
	n.restartElectionTimer()
	n.campaigner.Campaign(n)
}

func (n *node) didIGetMajority() bool {
	peers := n.whoArePeers.All()
	return n.st.votesGot >= (len(peers)/2 + 1)
}

func (n *node) gotVote(evt event) {
	if n.st.mode != Candidate {
		// must be a delayed vote by a node, ignore it
		return
	}
	voteResponse := evt.payload.(*voteResponse)
	if voteResponse.success {
		n.handleSuccessfulVoteResponse(voteResponse)
	} else {
		n.handleRejectedVoteResponse(voteResponse)
	}
}

func (n *node) handleRejectedVoteResponse(voteResponse *voteResponse) {
	term := getCurrentTerm(n)
	if term < voteResponse.term {
		n.dispatcher.Dispatch(event{StepDown, voteResponse.term})
	}
}

func (n *node) stepDown(evt event) {
	newTerm := evt.payload.(uint64)
	n.store.StoreInt(CurrentTermKey, newTerm)
	n.st.mode = Follower
	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) handleSuccessfulVoteResponse(*voteResponse) {
	n.st.votesGot = n.st.votesGot + 1
	if n.didIGetMajority() {
		n.st.mode = Leader
		n.dispatcher.Dispatch(event{StartLeader, nil})
	}
}

func (n *node) haveIAlreadyVotedForAnotherPeerForTheTerm(term uint64, peerAskingForVote string) bool {
	//NOTE Ref: notes 3.
	candidateVotedFor := n.votedFor.Get(term)
	if candidateVotedFor == "" || candidateVotedFor == peerAskingForVote {
		return false
	}
	return true
}

func (n *node) isCandidatesLogUptoDate(voteRequest *voteRequest) bool {
	// NOTE: Ref Notes 1.
	if voteRequest.lastLogTerm > n.log.LastTerm() {
		return true
	} else if (voteRequest.lastLogTerm == n.log.LastTerm()) && (voteRequest.lastLogIndex >= n.log.LastIndex()) {
		return true
	}
	return false
}

func (n *node) checkIfHigherTerm(termGot uint64) {
	term := getCurrentTerm(n)
	if termGot > term {
		if n.st.mode != Follower {
			n.dispatcher.Dispatch(event{StepDown, termGot})
		} else {
			n.store.StoreInt(CurrentTermKey, termGot)
		}
	}
}

func (n *node) gotRequestForVote(evt event) {
	voteRequest := evt.payload.(*voteRequest)
	term := getCurrentTerm(n)
	if voteRequest.term < term {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
	n.checkIfHigherTerm(voteRequest.term)
	if n.haveIAlreadyVotedForAnotherPeerForTheTerm(voteRequest.term, voteRequest.from.id) {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
	if !n.isCandidatesLogUptoDate(voteRequest) {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
	n.votedFor.Store(voteRequest.term, voteRequest.from.id)
	n.transport.SendVoteResponse(voteRequest.from, voteResponse{true, term, peer{n.id}})
}

func (n *node) appendToLog(request *appendEntriesRequest) {
	logIndex := request.prevLogIndex
	for _, appendEntry := range request.entries {
		logEntry, ok := n.log.EntryAt(logIndex)
		if ok && logEntry.term != appendEntry.term {
			n.log.DeleteFrom(logIndex)
		}
		n.log.AddAt(logIndex, appendEntry)
		logIndex = logIndex + 1
	}
}

func (n *node) appendEntries(evt event) {
	request := evt.payload.(*appendEntriesRequest)
	term := getCurrentTerm(n)
	if request.term < term {
		n.transport.SendAppendEntryResponse(request.from, appendEntriesResponse{false, term})
		return
	}
	// we heard from the leader here: all the other checks below are for log matching, think about this later!!!
	n.st.lastHeardFromALeader = n.time.UnixNow()
	// Important note: Reference 2
	if n.st.mode == Candidate {
		n.dispatcher.Dispatch(event{StepDown, request.term})
	} else {
		n.checkIfHigherTerm(request.term)
	}
	entry, ok := n.log.EntryAt(request.prevLogIndex)
	if ok && entry.term != request.prevLogTerm {
		n.transport.SendAppendEntryResponse(request.from, appendEntriesResponse{false, term})
		return
	}
	n.appendToLog(request)
	if request.leaderCommit > n.st.commitIndex {
		n.st.commitIndex = min(request.leaderCommit, n.log.LastIndex())
	}
}
func (n *node) sendHeartbeat() {
	term := getCurrentTerm(n)
	peers := n.whoArePeers.All()
	n.transport.SendAppendEntriesRequest(peers,
		appendEntriesRequest{from: peer{n.id}, term: term, prevLogTerm: n.log.LastTerm(),
			prevLogIndex: n.log.LastIndex(), entries: nil, leaderCommit: n.st.commitIndex})
	n.st.lastSentAppendEntriesAt = n.time.UnixNow()
}

func (n *node) startLeader(evt event) {
	if n.st.mode != Leader {
		panic("startLeader invoked when mode is not set as leader")
	}
	n.sendHeartbeat()
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) heartbeatTimerTimedout(evt event) {
	if n.st.mode != Leader {
		panic("Received heartbeat timer timedout when mode is not set as leader")
	}
	if (n.time.UnixNow() - n.st.lastSentAppendEntriesAt) < timeBetweenHeartbeats {
		n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
		return
	}
	n.sendHeartbeat()
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) gotAppendEntriesResponse(evt event) {
	appendEntriesResponse := evt.payload.(*appendEntriesResponse)
	if appendEntriesResponse.success && n.st.mode == Leader {
		// process success response here
	} else {
		n.checkIfHigherTerm(appendEntriesResponse.term)
	}
}
