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
	d := newDirectDispatcher()
	if mockTickerFn == nil {
		mockTickerFn = getMockTickerFn
	}
	if peers == nil {
		peers = []peer{}
	}

	return NewNodeWithDI("node-1", depends{dispatcher: d, store: store, getTicker: mockTickerFn, peersExplorer: newSimplePeersExplorer(peers), campaigner: noCampaigner})
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
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

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
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

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
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	c := n.st.stFn.(*candidate)
	if c.votesReceived != 1 {
		t.Fatal("candidate did not vote for itself")
	}
}

func Test_OnTransitionToACandidateItShouldAskForVotesFromPeers(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})
	var actualPeers []peer
	n.d.campaigner = func(c *node) func(peers []peer, currentTerm uint64) {
		return func(peers []peer, currentTerm uint64) {
			actualPeers = peers
		}
	}

	// trigger the election signal
	mockTicker.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	if actualPeers == nil || len(actualPeers) != 1 || actualPeers[0].Id != "peer1" {
		t.Fatal("peers not asked for vote")
	}
}

func Test_AsACandiateOnElectionSignalTransitionsToAFollower(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})

	n.d.campaigner = func(c *node) func(peers []peer, currentTerm uint64) {
		return func(peers []peer, currentTerm uint64) {
		}
	}

	// trigger the election signal
	mockTicker.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	// trigger the election signal
	mockTicker.tick()
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have transitioned to a follower, after getting election signal as a candidate")
	}

}


func Test_AsACandiateOnGettingARejectedVoteTransitionsToAFollower(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})

	n.d.campaigner = func(c *node) func(peers []peer, currentTerm uint64) {
		return func(peers []peer, currentTerm uint64) {
		}
	}

	// trigger the election signal
	mockTicker.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}
	term, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		t.Fatal("Could not get current term")
	}
	directD.dispatch(event{evtType:GotVoteRequestRejected,st:n.st,payload:&voteResponse{Success:false,Term:(term+1),From:"peer1"}})
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have transitioned to a follower, after getting a rejected vote as a candidate")
	}

}

func Test_AsACandiateOnGettingRequisiteVotesTransitionsToALeader(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, []peer{{"peer0","address0"},{"peer1", "address1"}})

	n.d.campaigner = func(c *node) func(peers []peer, currentTerm uint64) {
		return func(peers []peer, currentTerm uint64) {
		}
	}

	// trigger the election signal
	mockTicker.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}
	term, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		t.Fatal("Could not get current term")
	}
	directD.dispatch(event{evtType:GotVote,st:n.st,payload:&voteResponse{Success:true,Term:(term),From:"peer1"}})
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*leader)(nil)) {
		t.Fatal("Should have transitioned to a leader, after getting the required number of votes as a candidate")
	}

}


func Test_AsACandiateGetsLessThanMajorityVotesDoesNotGetElectedAsLeader(t *testing.T) {
	var mockTicker = newMockTicker(time.Duration(1))
	var g getTickerFn = func(d time.Duration) Ticker {
		return mockTicker
	}

	n := getNode(g, nil, []peer{ {"peer0", "address0"},{"peer1", "address1"},{"peer2","address2"},{"peer3","address3"} })

	n.d.campaigner = func(c *node) func(peers []peer, currentTerm uint64) {
		return func(peers []peer, currentTerm uint64) {
		}
	}

	// trigger the election signal
	mockTicker.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	term, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		t.Fatal("Could not get current term")
	}

	// a no from peer1 and peer2
	directD.dispatch(event{evtType:GotVoteRequestRejected,st:n.st,payload:&voteResponse{Success:false,Term:(term),From:"peer1"}})
	directD.awaitSignal()
	directD.reset()

	directD.dispatch(event{evtType:GotVoteRequestRejected,st:n.st,payload:&voteResponse{Success:false,Term:(term),From:"peer2"}})
	directD.awaitSignal()
	directD.reset()

	// a yes from peer3
	directD.dispatch(event{evtType:GotVote,st:n.st,payload:&voteResponse{Success:true,Term:(term),From:"peer3"}})
	directD.awaitSignal()
	directD.reset()


	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should still have remained as a candidate, after getting less than majority votes")
	}

}
