package raft

// TO DO:
// If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)
import (
	"fmt"
	"log"
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
	n.st.matchIndex = make(map[string]uint64)
	n.st.nextIndex = make(map[string]uint64)
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
	case GotCommand:
		n.gotCommand(event)
	default:
		log.Panic(fmt.Sprintf("Unknown event: %d passed to handleEvent", event.eventType))
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
		log.Panic("Mode is not set to follower in startFollower")
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
		// The following will have to be changed to check if the leader
		// is still able to reach majority of the nodes within an election time out
		log.Panic("Received election timer timedout while being a leader")
	}
}

func (n *node) startCandidate(event event) {
	if n.st.mode != Candidate {
		log.Panic("Mode is not set to candidate in startCandidate")
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
	voteResponse := evt.payload.(*voteResponse)
	if n.isHigherTerm(voteResponse.term) {
		n.handleHigherTermReceived(voteResponse.term)

	} else if voteResponse.success {
		n.handleSuccessfulVoteResponse(voteResponse)
	}
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

func (n *node) isHigherTerm(termGot uint64) bool {
	term := getCurrentTerm(n)
	return termGot > term
}

func (n *node) handleHigherTermReceived(termGot uint64) {
	n.store.StoreInt(CurrentTermKey, termGot)
	if n.st.mode != Follower {
		n.st.mode = Follower
		n.dispatcher.Dispatch(event{StartFollower, nil})
	}
}

func (n *node) decideVoteResponse(voteRequest *voteRequest, term uint64) voteResponse {
	var response voteResponse
	if n.haveIAlreadyVotedForAnotherPeerForTheTerm(voteRequest.term, voteRequest.from.id) {
		response = voteResponse{false, term, peer{n.id}}
	} else if !n.isCandidatesLogUptoDate(voteRequest) {
		response = voteResponse{false, term, peer{n.id}}
	} else {
		response = voteResponse{true, term, peer{n.id}}
	}
	return response
}

func (n *node) gotRequestForVote(evt event) {
	voteRequest := evt.payload.(*voteRequest)
	term := getCurrentTerm(n)
	if voteRequest.term < term {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
	response := n.decideVoteResponse(voteRequest, term)
	if response.success {
		n.votedFor.Store(voteRequest.term, voteRequest.from.id)
	}
	n.transport.SendVoteResponse(voteRequest.from, response)
	if n.isHigherTerm(voteRequest.term) {
		n.handleHigherTermReceived(voteRequest.term)
	}
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

func (n *node) stepDown(term uint64) {
	n.st.mode = Follower
	n.store.StoreInt(CurrentTermKey, term)
	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) appendEntries(evt event) {
	request := evt.payload.(*appendEntriesRequest)
	term := getCurrentTerm(n)
	if request.term < term {
		n.transport.SendAppendEntriesResponse(request.from, appendEntriesResponse{success: false, term: term, from: n.id})
		return
	}
	// we heard from the leader here: all the other checks below are for log matching, think about this later!!!
	n.st.lastHeardFromALeader = n.time.UnixNow()
	response := n.ifOkAppendToLog(request, term)
	n.transport.SendAppendEntriesResponse(request.from, response)
	// Important note: Reference 2
	if n.isHigherTerm(request.term) {
		n.handleHigherTermReceived(request.term)
	} else if n.st.mode == Candidate {
		n.dispatcher.Dispatch(event{StartFollower, nil})
	}
}

func (n *node) ifOkAppendToLog(request *appendEntriesRequest, term uint64) appendEntriesResponse {
	entry, ok := n.log.EntryAt(request.prevLogIndex)
	if ok && entry.term != request.prevLogTerm {
		return appendEntriesResponse{success: false, term: term, from: n.id}
	} else if !ok && request.prevLogIndex != 0 {
		// request.prevLogIndex == 0 ==> implies that the leader, when sending the first log entry at index 1 sends prevLogIndex = 0
		return appendEntriesResponse{success: false, term: term, from: n.id}
	}
	n.appendToLog(request)
	if request.leaderCommit > n.st.commitIndex {
		n.st.commitIndex = min(request.leaderCommit, n.log.LastIndex())
	}
	return appendEntriesResponse{success: true, term: term, from: n.id}
}

func (n *node) sendAppendEntries(isHeartbeat bool) {
	term := getCurrentTerm(n)
	peers := n.whoArePeers.All()
	for _, p := range peers {
		entries := n.getEntriesToReplicate(p)

		if (entries == nil || len(entries) == 0) && isHeartbeat {
			n.transport.SendAppendEntriesRequest(p,
				appendEntriesRequest{from: peer{n.id}, term: term, prevLogTerm: n.log.LastTerm(),
					prevLogIndex: n.log.LastIndex(), entries: nil, leaderCommit: n.st.commitIndex})
		} else {
			//Reference: Notes: point 7
			n.transport.SendAppendEntriesRequest(p,
				appendEntriesRequest{from: peer{n.id}, term: term, prevLogTerm: n.log.LastTerm(),
					prevLogIndex: n.log.LastIndex(), entries: entries, leaderCommit: n.st.commitIndex})
		}
		n.st.lastSentAppendEntriesAt = n.time.UnixNow()
	}
}

func (n *node) getEntriesToReplicate(p peer) []entry {
	nextIndex := n.st.nextIndex[p.id]
	entries := n.log.Get(nextIndex)
	return entries
}

func (n *node) initializeLeaderState() {
	lastLogIndex := n.log.LastIndex()
	peers := n.whoArePeers.All()
	for _, peer := range peers {
		n.st.matchIndex[peer.id] = 0
		n.st.nextIndex[peer.id] = lastLogIndex + 1
	}
}

func (n *node) startLeader(evt event) {
	if n.st.mode != Leader {
		log.Panic("startLeader invoked when mode is not set as leader")
	}
	n.initializeLeaderState()
	n.sendAppendEntries(true)
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) heartbeatTimerTimedout(evt event) {
	if n.st.mode != Leader {
		log.Panic("Received heartbeat timer timedout when mode is not set as leader")
	}

	if (n.time.UnixNow() - n.st.lastSentAppendEntriesAt) > timeBetweenHeartbeats {
		n.sendAppendEntries(true)
	}
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) gotCommand(evt event) {
	if n.st.mode != Leader {
		// not a leader, redirect to leader later
		log.Panic("Not a leader, received command. Method - yet to be implemented")
		return
	}
	command := evt.payload.(string)
	currentTerm := getCurrentTerm(n)
	entry := entry{term: currentTerm, command: command}
	n.log.Append(entry)
	n.sendAppendEntries(false)
}

func onAppendEntriesSucceeded(response *appendEntriesResponse) {
	log.Panic("Implement this!")
}

func (n *node) gotAppendEntriesResponse(evt event) {
	if n.st.mode != Leader {
		// what is the ideal way to handle this? does this situation occur?
		return
	}
	response := evt.payload.(*appendEntriesResponse)
	if response.success {
		onAppendEntriesSucceeded(response)
	} else if n.isHigherTerm(response.term) {
		n.handleHigherTermReceived(response.term)
	} else {
		// we need to decrement the nextIndex and send append entries to the peer
		peerId := response.from
		value, ok := n.st.nextIndex[peerId]
		if ok {
			n.st.nextIndex[peerId] = value - 1
		} else {
			log.Print("Unknown peer")
		}
	}
	//Reference: Notes point 8
}
