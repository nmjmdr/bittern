package raft

import (
	"reflect"
	"testing"
	"time"
)

func Test_OnTransitionToACandidateItShouldVoteForItself(t *testing.T) {
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, nil)

	// trigger the election signal
	mockTimer.tick()
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
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})
	var actualPeers []peer
	n.d.chatter.(*mockChatter).campaignStub = func(peers []peer, currentTerm uint64) {
		actualPeers = peers
	}

	// trigger the election signal
	mockTimer.tick()
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
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})

	// trigger the election signal
	mockTimer.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	// trigger the election signal
	mockTimer.tick()
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have transitioned to a follower, after getting election signal as a candidate")
	}

}

func Test_AsACandiateOnGettingARejectedVoteTransitionsToAFollower(t *testing.T) {
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer1", "address1"}})

	// trigger the election signal
	mockTimer.tick()
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
	directD.dispatch(event{evtType: GotVoteRequestRejected, st: n.st, payload: &voteResponse{success: false, term: (term + 1), from: "peer1"}})
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have transitioned to a follower, after getting a rejected vote as a candidate")
	}
}

func Test_AsACandiateOnGettingRequisiteVotesTransitionsToALeader(t *testing.T) {
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer0", "address0"}, {"peer1", "address1"}})

	// trigger the election signal
	mockTimer.tick()
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
	directD.dispatch(event{evtType: GotVote, st: n.st, payload: &voteResponse{success: true, term: (term), from: "peer1"}})
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*leader)(nil)) {
		t.Fatal("Should have transitioned to a leader, after getting the required number of votes as a candidate")
	}
}

func Test_AsACandiateGetsLessThanMajorityVotesDoesNotGetElectedAsLeader(t *testing.T) {
	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer0", "address0"}, {"peer1", "address1"}, {"peer2", "address2"}, {"peer3", "address3"}})

	// trigger the election signal
	mockTimer.tick()
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
	directD.dispatch(event{evtType: GotVoteRequestRejected, st: n.st, payload: &voteResponse{success: false, term: (term), from: "peer1"}})
	directD.awaitSignal()
	directD.reset()

	directD.dispatch(event{evtType: GotVoteRequestRejected, st: n.st, payload: &voteResponse{success: false, term: (term), from: "peer2"}})
	directD.awaitSignal()
	directD.reset()

	// a yes from peer3
	directD.dispatch(event{evtType: GotVote, st: n.st, payload: &voteResponse{success: true, term: (term), from: "peer3"}})
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should still have remained as a candidate, after getting less than majority votes")
	}
}

func Test_AsACandiateWhenItGetsAppendEntryItTransitionsToAFollower(t *testing.T) {

	var mockTimer = newMockTimer(time.Duration(1))
	var g getTimerFn = func(d time.Duration) Timer {
		return mockTimer
	}

	n := getNode(g, nil, []peer{{"peer0", "address0"}, {"peer1", "address1"}, {"peer2", "address2"}, {"peer3", "address3"}})
	n.d.time = newMockTime(electionTimeSpan + 100)
	var er appendEntryResponse
	n.d.chatter.(*mockChatter).sendAppendEntryResponseStub = func(response appendEntryResponse) {
		er = response
	}

	// trigger the election signal
	mockTimer.tick()
	directD := n.d.dispatcher.(*directDispatcher)
	//<- directD.dispatched
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*candidate)(nil)) {
		t.Fatal("Should have been a candidate, after getting election signal")
	}

	currentTerm, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		t.Fatal("Not able to get current term")
	}

	evt := event{evtType: AppendEntry, st: n.st, payload: &appendEntryRequest{term: currentTerm}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should have transitioned to a follower, as it did hear from a leader")
	}

	if !er.success {
		t.Fatal("Should NOT have rejected the append entry request")
	}
}
