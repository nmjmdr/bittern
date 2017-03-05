package raft

import (
	"testing"
)

func createNode() *node {
	n := newNode()
	n.dispatcher = newMockDispathcer(nil)
  n.store = newInMemoryStore()
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
  currentTerm, ok := n.store.getInt(CurrentTermKey)
	if !ok || currentTerm != 0 {
		t.Fatal("Should have initialized the term to 0")
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
