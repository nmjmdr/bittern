package raft

import (
  "fmt"
  "time"
  "math/rand"
)

const CurrentTermKey = "current-term"

type node struct {
	st         *state
	dispatcher Dispatcher
  store      Store
  followerExpiryTimer Timer
}

func newNode() *node {
  rand.Seed(time.Now().Unix())
	n := new(node)
	return n
}

func (n *node) boot() {
	n.st = new(state)
	n.store.StoreInt(CurrentTermKey,0)
	n.st.mode = Follower
  n.st.commitIndex = 0
  n.st.lastApplied = 0

	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) handleEvent(event event) {
  switch(event.eventType) {
  case StartFollower:
    n.startFollower(event)
  default:
    panic(fmt.Sprintf("Unknown event: %d passed to handleEvent",event.eventType))
  }
}

func (n *node) startFollower(event event) {
  if n.st.mode != Follower {
    panic("Mode is not a follower in startFollower")
  }
  n.followerExpiryTimer.Start(getRandomizedElectionTimout())
}
