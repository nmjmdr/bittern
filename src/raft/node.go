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

	dispatcher          Dispatcher
	store               Store
	electionExpiryTimer ElectionTimeoutTimer
	time                Time
	campaigner          Campaigner
	whoArePeers         WhoArePeers
}

func newNode() *node {
	rand.Seed(time.Now().Unix())
	n := new(node)
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
	term, ok := n.store.GetInt(CurrentTermKey)
	if !ok {
		panic("Not able to to obtain current term in startCandidate")
	}
	term = term + 1
	n.store.StoreInt(CurrentTermKey, term)
	n.st.votesGot = n.st.votesGot + 1
	n.restartElectionTimer()
	n.campaigner.Campaign(n)
}

func (n *node) gotMajority() bool {
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
	term, ok := n.store.GetInt(CurrentTermKey)
	if !ok {
		panic("Not able to to obtain current term in handleRejectedVoteResponse")
	}
	if term < voteResponse.term {
		n.dispatcher.Dispatch(event{StepDown, term})
	}
}

func (n *node) handleSuccessfulVoteResponse(*voteResponse) {
	n.st.votesGot = n.st.votesGot + 1
	if n.gotMajority() {
		n.st.mode = Leader
		// what else needs to be done here?
		n.dispatcher.Dispatch(event{StartLeader, nil})
	}
}
