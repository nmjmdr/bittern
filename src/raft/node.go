package raft

import (
	"fmt"
	"math/rand"
	"time"
)

const CurrentTermKey = "current-term"

type node struct {
	st              *state
	electionTimeout int64
	id              string

	dispatcher          Dispatcher
	store               Store
	electionExpiryTimer ElectionTimeoutTimer
	time                Time
	campaigner          Campaigner
	whoArePeers         WhoArePeers
	transport           Transport
	votedFor            VotedFor
	log                 Log
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
		n.dispatcher.Dispatch(event{StepDown, term})
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
	// have we already voted for another peer in request's term?
	// If votedFor is empty or candidateId, then grant vote
	// where votedFor => candidateId that received vote in current term (or null if none)
	// who did we vote for in the term?
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

func (n *node) gotRequestForVote(evt event) {
	voteRequest := evt.payload.(*voteRequest)
	term := getCurrentTerm(n)
	if voteRequest.term < term {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
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
	entry, ok := n.log.EntryAt(request.prevLogIndex)
	if ok && entry.term != request.prevLogTerm {
		n.transport.SendAppendEntryResponse(request.from, appendEntriesResponse{false, term})
		return
	}
	n.appendToLog(request)
	if request.leaderCommit > n.st.commitIndex {
		n.st.commitIndex = min(request.leaderCommit, n.log.LastIndex())
	}
	// Important note: Reference 2
	if n.st.mode == Candidate {
		n.dispatcher.Dispatch(event{StepDown, term})
	}
}

func (n *node) startLeader(evt event) {
	term := getCurrentTerm(n)
	peers := n.whoArePeers.All()
	n.transport.SendAppendEntriesRequest(peers,
		appendEntriesRequest{from: peer{n.id}, term: term, prevLogTerm: n.log.LastTerm(),
			prevLogIndex: n.log.LastIndex(), entries: nil, leaderCommit: n.st.commitIndex})

}
