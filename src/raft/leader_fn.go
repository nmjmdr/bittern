package raft


type leader struct {
	*node
}

func newLeader(n *node) *leader {
	l := new(leader)
	l.node = n
	return l
}

func (l *leader) gotElectionSignal() {
  // ?
}

func (l *leader) gotVote(evt event) {
  // ignore, possibly a slow client
}

func (l *leader) gotVoteRequestRejected(evt event) {
	// ignore, possibly a slow client
}

func (l *leader) gotRequestForVote(evt event) {
  // process it, it might be that the leader's append entries are not reaching other clients
	// if the candidate gets elected, the leader steps down
	respondToVoteRequest(evt,l.node)
}
