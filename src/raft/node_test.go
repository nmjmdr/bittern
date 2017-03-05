package raft

import (
	"testing"
  "time"
)

func createNode() *node {
	n := newNode()
	n.dispatcher = newMockDispathcer(nil)
	n.store = newInMemoryStore()
	n.followerExpiryTimer = newMockElectionTimeoutTimer(nil)
	return n
}

func Test_when_the_node_boots_it_should_start_as_a_follower(t *testing.T) {
	n := createNode()
	n.boot()
	if n.st.mode != Follower {
		t.Fatal("Should have initialized as a follower")
	}
}

func Test_when_the_node_boots_it_should_set_the_term_to_zero(t *testing.T) {
	n := createNode()
	n.boot()
	currentTerm, ok := n.store.GetInt(CurrentTermKey)
	if !ok || currentTerm != 0 {
		t.Fatal("Should have initialized current term to 0")
	}
}

func Test_when_the_node_boots_it_should_set_the_commit_index_to_zero(t *testing.T) {
	n := createNode()
	n.boot()
	if n.st.commitIndex != 0 {
		t.Fatal("Should have initialized commit index to 0")
	}
}

func Test_when_the_node_boots_it_should_set_the_last_applied_to_zero(t *testing.T) {
	n := createNode()
	n.boot()
	if n.st.lastApplied != 0 {
		t.Fatal("Should have initialized last applied to 0")
	}
}

func Test_when_the_node_boots_it_should_generate_start_follower_event(t *testing.T) {
	n := createNode()
	var gotEvent event
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		gotEvent = event
	}
	n.boot()
	if gotEvent.eventType != StartFollower {
		t.Fatal("Should have generated start follower event")
	}
}

func Test_when_start_follower_event_is_handled_it_should_start_the_election_timeout_countdown(t *testing.T) {
	n := createNode()
	timerStarted := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	n.followerExpiryTimer.(*mockElectionTimeoutTimer).callback = func(t time.Duration) {
		timerStarted = true
	}
	n.boot()
	if !timerStarted {
		t.Fatal("Election timeout timer not started")
	}
}
