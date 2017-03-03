package raft

import (
	"reflect"
	"testing"
)

func Test_onInitItShouldSetCommitIndexTo0(t *testing.T) {
	n := getNode(nil, nil, nil)
	if n.st.commitIndex != 0 {
		t.Fatal("Commit Index should have been intialized to 0")
	}
}

func Test_onInitItShouldSetLastAppliedTo0(t *testing.T) {
	n := getNode(nil, nil, nil)
	if n.st.lastApplied != 0 {
		t.Fatal("Commit Index should have been intialized to 0")
	}
}

func Test_onInitItShouldStartAsAFollower(t *testing.T) {
	n := getNode(nil, nil, nil)
	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have been a follower")
	}
}

func Test_CanGetCurrentIndexFromstore(t *testing.T) {
	store := newInMemorystore()
	expected := uint64(100)
	store.setInt("current-term", expected)
	n := getNode(nil, store, nil)
	actual, ok := n.d.store.getInt("current-term")
	if !ok {
		t.Fatal("could not obtain current-term from store")
	}

	if actual != expected {
		t.Fatal("Could not get the expected value for current-index")
	}
}

func Test_CanGetVotedForFromStore(t *testing.T) {
	store := newInMemorystore()
	expected := "node-2"
	store.setValue("voted-for", expected)
	n := getNode(nil, store, nil)
	actual, ok := n.d.store.getValue("voted-for")
	if !ok {
		t.Fatal("could not obtain voted-for from store")
	}

	if actual != expected {
		t.Fatal("Could not get the expected value for voted-for")
	}
}

func Test_OnBootGetsCurrentTermAs0(t *testing.T) {
	n := getNode(nil, nil, nil)
	actual, ok := n.d.store.getInt(currentTermKey)

	if !ok {
		t.Fatal("On boot should have set the value of current-term to 0")
	}

	if actual != 0 {
		t.Fatal("On boot should have set the value of current-term to 0")
	}
}

func Test_OnInitStartsThedispatcher(t *testing.T) {
	store := newInMemorystore()
	d := &directDispatcher{}
	NewNodeWithDI("node-1", depends{dispatcher: d, store: store, getTimer: getMockTimerFn})

	if d.started == false {
		t.Fatal("dispatcher not started")
	}
}
