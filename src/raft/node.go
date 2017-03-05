package raft

import (
	"fmt"
	"math/rand"
	"time"
)

const CurrentTermKey = "current-term"

type node struct {
	st                   *state
	dispatcher           Dispatcher
	store                Store
	followerExpiryTimer  ElectionTimeoutTimer
	lastHeardFromALeader int64
  electionTimeout int64
}

func newNode() *node {
	rand.Seed(time.Now().Unix())
	n := new(node)
	return n
}

func (n *node) boot() {
	n.st = new(state)
	n.store.StoreInt(CurrentTermKey, 0)
	n.st.mode = Follower
	n.st.commitIndex = 0
	n.st.lastApplied = 0

	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) handleEvent(event event) {
	switch event.eventType {
	case StartFollower:
		n.startFollower(event)
  case ElectionTimerTimedout:
    n.electionTimerTimeout(event)
	default:
		panic(fmt.Sprintf("Unknown event: %d passed to handleEvent", event.eventType))
	}
}

func (n *node) startFollower(event event) {
	if n.st.mode != Follower {
		panic("Mode is not set to follower in startFollower")
	}
  n.electionTimeout = getRandomizedElectionTimout()
	n.followerExpiryTimer.Start(time.Duration(n.electionTimeout) * time.Millisecond)
}

func (n *node) hasHeardFromALeader() bool {
  return (time.Now().Unix() - n.lastHeardFromALeader) < n.electionTimeout
}

func (n *node) electionTimerTimeout(event event) {
  if n.hasHeardFromALeader() {
    return
  }

  // TO DO: transition to a candidate
  
}
