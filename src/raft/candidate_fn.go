package raft

import (
	"fmt"
)

type candidate struct {
	*node
	votesReceived int
}

func newCandidate(n *node) *candidate {

	c := new(candidate)
	c.node = n
	// vote for self
	c.votesReceived = c.votesReceived + 1
	// begins election timer
	beginElectionTimer(c.d.getTicker,c.d.dispatcher,c.st)
	// TO DO: start requesting vote from peers
	peers := c.d.peersExplorer.getPeers()
	currentTerm := getCurrentTerm(c.d.store)
	c.d.chatter.campaign(peers, currentTerm)
	return c
}

func (c *candidate) gotElectionSignal() {
	// transition to a follower. The follower restrats the election timer
	c.st.stFn = newFollower(c.node)
}

func (c *candidate) gotVote(evt event) {
	// TO DO
	// TO DO: increment the vote count
	// check if we have necessary amount of votes?
	// if yes transition to leader
	peers := c.d.peersExplorer.getPeers()
	c.votesReceived = c.votesReceived + 1
	// have received a majority of the votes
	if c.votesReceived >= ((len(peers)/2) + 1) {
		// transition to a leader
		c.st.stFn = newLeader(c.node)
	}

}

func (c *candidate) gotVoteRequestRejected(evt event) {
	// got a rejected vote, could be that the peer has already voted
	// or could be a higher term, check if we have discovered a higher term
	currentTerm, ok := c.d.store.getInt(currentTermKey)
	if !ok {
		panic("Could not obtain current term key in candidate")
	}
	vr := evt.payload.(*voteResponse)
	if currentTerm < vr.Term {
		// transition to a follower
		c.st.stFn = newFollower(c.node)
	}
}

func checkCandidatesLog() bool {
	fmt.Println("Not checking candidate's log!!! check it")
	return true
}

func (c *candidate) gotRequestForVote(evt event) {
	request := evt.payload.(*voteRequest)
	currentTerm, ok := c.d.store.getInt(currentTermKey)
	if !ok {
		panic("Could not obtain current term key in candidate")
	}
	if request.term < currentTerm {
		// reject
		c.d.chatter.sendVoteResponse(voteResponse{Success:false,Term:currentTerm,From:c.id})
		return
	}
	votedFor, ok := c.d.store.getValue(votedForKey)
	if ok && votedFor != request.from  && checkCandidatesLog() {
		// grant vote
		c.d.chatter.sendVoteResponse(voteResponse{Success:true,Term:currentTerm,From:c.id})
		// save votedFor
		c.d.store.setValue(votedForKey,request.from)
		// do not transition to a follower, it might so happen that,this candidate might ending up
		// geting required number of votes
		return
	}

	c.d.chatter.sendVoteResponse(voteResponse{Success:false,Term:currentTerm,From:c.id})
	return
}
