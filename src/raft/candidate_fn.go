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
	return c
}

func (c *candidate) gotElectionSignal() {
	// transition to a follower. The follower restrats the election timer
	fmt.Println("here in got election signal")
	c.st.stFn = newFollower(c.node)
}
