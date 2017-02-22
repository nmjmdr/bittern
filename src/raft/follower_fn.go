package raft

import (
	//"fmt"
)

type follower struct {
	*node
}

func newFollower(n *node) *follower {
	f := new(follower)
	f.node = n
	beginElectionTimer(f.d.getTicker,f.d.dispatcher,f.st)
	return f
}

func (f *follower) gotElectionSignal() {
	// we do not need worry about whether
	// a. Should we increment the term and then transition to a canidate
	// or b. should we transition to a candidate and then increment the term
	// We need not worry because all events are serialized onto the queue, the next event
	// will be handled only after the node transitions to a candidate

	// increment the current term
	currentTerm, ok := f.node.d.store.getInt(currentTermKey)
	if !ok {
		panic("Unable to obtain current term from store")
	}
	currentTerm = currentTerm + 1
	f.d.store.setInt(currentTermKey, currentTerm)

	//transition to a candidate
	f.st.stFn = newCandidate(f.node)
}

func (f *follower) gotVote(evt event) {
	// already a follower, we must be getting this message from
	// another node which is out of sync with the restrats
	// ignore
}

func (f *follower) gotVoteRequestRejected(evt event) {
	// already a a follower, probably a delayed response by a node
	// ignore
}


func (f *follower) gotRequestForVote(evt event) {

	request := evt.payload.(*voteRequest)
	currentTerm, ok := f.d.store.getInt(currentTermKey)
	if !ok {
		panic("Could not obtain current term key in candidate")
	}
	if request.term < currentTerm {
		// reject
		f.d.chatter.sendVoteResponse(voteResponse{Success:false,Term:currentTerm,From:f.id})
		return
	}

	votedFor, ok := f.d.store.getValue(votedForKey)

	if (ok && votedFor == request.from) || !checkCandidatesLog() {
		f.d.chatter.sendVoteResponse(voteResponse{Success:false,Term:currentTerm,From:f.id})
		return
	}


		// grant vote
	f.d.chatter.sendVoteResponse(voteResponse{Success:true,Term:currentTerm,From:f.id})

}
