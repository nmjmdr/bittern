package raft

const CurrentTermKey = "current-term"

type node struct {
	st         *state
	dispatcher Dispatcher
  store      Store
}

func newNode() *node {
	n := new(node)
	return n
}

func (n *node) boot() {
	n.st = new(state)
	n.store.storeInt(CurrentTermKey,0)
	n.st.mode = Follower

	n.dispatcher.Dispatch(event{StartFollower, nil})
}
