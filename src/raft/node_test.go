package raft

import (
	"reflect"
	"testing"
	"time"
)

var getMockTickerFn getTickerFn = func(d time.Duration) Ticker {
	return newMockTicker(d)
}

func getNode(mockTickerFn getTickerFn, store store, peers []peer) *node {

	if store == nil {
		store = newInMemorystore()
	}
	d := &directDispatcher{}
	if mockTickerFn == nil {
		mockTickerFn = getMockTickerFn
	}
	if peers == nil {
		peers = []peer{}
	}

	return NewNodeWithDI("node-1", depends{dispatcher: d, store: store, getTicker: mockTickerFn, transport: newMockTransport(), peersExplorer: newSimplePeersExplorer(peers)})
}

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

func Test_CanGetVotedForFromstore(t *testing.T) {
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
	NewNodeWithDI("node-1", depends{dispatcher: d, store: store, getTicker: getMockTickerFn})

	if d.started == false {
		t.Fatal("dispatcher not started")
	}
}

func Test_AsAFollowerStartsTheElectionTimer(t *testing.T) {
	called := false
	var g getTickerFn = func(d time.Duration) Ticker {
		called = true
		return newMockTicker(d)
	}

	getNode(g, nil, nil)

	if called == false {
		t.Fatal("Mock ticker for election timer not initialized")
	}
}

func Test_OnElectionTSignalItShouldIncrementCurrentTerm(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, nil)

	// trigger the election signal
	mockTicker.tick()

	v, ok := n.d.store.getInt(currentTermKey)

	if !ok {
		t.Fatal("Unable to read current term key")
	}

	if v != 1 {
		t.Fatal("Current term should have incremented by 1")
	}
}

func Test_OnElectionTSignalItShouldTransitionToACandidate(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, nil)

	// trigger the election signal
	mockTicker.tick()
	// mock ticker does not run things concurrently, hence we can test for functionality
	// the same cannot be said about about event loop

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}
}

// Test if candidate votes for self
func Test_OnTransitionToACandidateItShouldVoteForItself(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, nil)

	// trigger the election signal
	mockTicker.tick()
	// mock ticker does not run things concurrently, hence we can test for functionality
	// the same cannot be said about about event loop

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	c := n.st.stFn.(*candidate)
	if c.votesReceived != 1 {
		t.Fatal("candidate did not vote for itself")
	}
}
