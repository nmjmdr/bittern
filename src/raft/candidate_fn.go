package raft

type candidate struct {
	*node
	votesReceived int
}

func newCandidate(n *node) *candidate {
	c := new(candidate)
	c.node = n
	// vote for self
	c.votesReceived = c.votesReceived + 1
	// start requesting vote from peers
	vr := voteRequest{}
	askVotes := getCampaign(c.node.d.transport, c.node.d.dispatcher, c.node.st,vr)
	askVotes(c.node.d.peersExplorer.getPeers())

	return c
}

func (c *candidate) gotElectionSignal(st *state) {
	// transition to a follower. The follower restrats the election timer
	st.stFn = newFollower(c.node)
}
