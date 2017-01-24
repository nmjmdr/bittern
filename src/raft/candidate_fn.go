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
	// TO DO: start requesting vote from peers
	peers := c.d.peersExplorer.getPeers();
	currentTerm, ok := c.d.store.getInt(currentTermKey)
	if !ok {
		panic("could not obtain current term as a candiate")
	}
	//c.campaign(peers,currentTerm)
	c.d.campaigner(c.node)(peers,currentTerm)
	return c
}

func (c *candidate) campaign(peers []peer,currentTerm uint64) {
	vReq := voteRequest{c.Id(),currentTerm}
	for _,peer := range peers {
		go func() {
			vRes, err := c.d.transport.askVote(peer,vReq)
			if err != nil {
				// log and return
				return
			}
			if vRes.Success {
				c.d.dispatcher.dispatch(event{GotVote,c.st,vRes})
			} else {
				c.d.dispatcher.dispatch(event{GotVoteRejected,c.st,vRes})
			}
		}()
	}
}

func (c *candidate) gotElectionSignal() {
	// transition to a follower. The follower restrats the election timer
	fmt.Println("here in got election signal")
	c.st.stFn = newFollower(c.node)
}
