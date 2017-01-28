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
	// begins election timer
	beginElectionTimer(c.d.getTicker,c.d.dispatcher,c.st)
	// TO DO: start requesting vote from peers
	peers := c.d.peersExplorer.getPeers()
	currentTerm := getCurrentTerm(c.d.store)
	c.d.campaigner(c.node)(peers, currentTerm)
	return c
}

func (c *candidate) gotElectionSignal() {
	// transition to a follower. The follower restrats the election timer
	c.st.stFn = newFollower(c.node)
}

func (c *candidate) gotVote(vr voteResponse) {
	// TO DO
	// TO DO: increment the vote count
	// check if we have necessary amount of votes?
	// if yes transition to leader
}

func (c *candidate) gotVoteRequestRejected(vr voteResponse) {
	// got a rejected vote, could be that the peer has already voted
	// or could be a higher term, check if we have discovered a higher term
	//currentTerm, ok :=
}
