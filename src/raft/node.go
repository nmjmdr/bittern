package raft

type node struct {
  st *state
  dispatcher Dispatcher
}

func newNode() *node {
  n := new(node)
  return n
}

func (n *node) boot() {
  n.st = new(state)
  n.st.term = 0
  n.st.mode = Follower

  n.dispatcher.Dispatch(event{StartFollower,nil})
}
