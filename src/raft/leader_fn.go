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
  // ?
}

func (l *leader) gotVoteRequestRejected(evt event) {
	//?
}

func (l *leader) gotRequestForVote(event event) {
  //?
}
