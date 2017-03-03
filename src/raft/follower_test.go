package raft

import (
	"reflect"
	"testing"
	"time"
)

func Test_AsAFollowerStartsTheElectionTimer(t *testing.T) {
	called := false
	var g getTimerFn = func(d time.Duration) Timer {
		called = true
		return newMockTimer(d)
	}

	getNode(g, nil, nil)

	if called == false {
		t.Fatal("Mock timer for election timer not initialized")
	}
}

func Test_AsAFollowerOnElectionSignalItShouldIncrementCurrentTerm(t *testing.T) {
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

	v, ok := n.d.store.getInt(currentTermKey)

	if !ok {
		t.Fatal("Unable to read current term key")
	}

	if v != 1 {
		t.Fatal("Current term should have incremented by 1")
	}
}

func Test_AsAFollowerOnElectionSignalItShouldTransitionToACandidate(t *testing.T) {
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
}

func Test_AsAFollowerVotesForACandidateWhenTheTermIsEqualToCurrentTerm(t *testing.T) {
	n := getNode(nil, nil, nil)
	currentTerm, ok := n.d.store.getInt(currentTermKey)

	if !ok {
		t.Fatal("Should have been able to read the current term")
	}

	var actualVoteResponse voteResponse
	n.d.chatter.(*mockChatter).sendVoteResponseStub = func(voteResponse voteResponse) {
		actualVoteResponse = voteResponse
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotRequestForVote, st: n.st, payload: &voteRequest{from: "1", term: currentTerm}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if !actualVoteResponse.success {
		t.Fatal("Should have got a successful vote response")
	}
}

func Test_AsAFollowerDoesVoteForACandidateWhenTheTermIsGreaterThanToCurrentTerm(t *testing.T) {
	n := getNode(nil, nil, nil)
	currentTerm, ok := n.d.store.getInt(currentTermKey)

	if !ok {
		t.Fatal("Should have been able to read the current term")
	}

	var actualVoteResponse voteResponse
	n.d.chatter.(*mockChatter).sendVoteResponseStub = func(voteResponse voteResponse) {
		actualVoteResponse = voteResponse
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotRequestForVote, st: n.st, payload: &voteRequest{from: "1", term: (currentTerm + 1)}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if !actualVoteResponse.success {
		t.Fatal("Should have got a successful vote response")
	}
}

func Test_AsAFollowerDoesNotVoteForACandidateWhenItHasAlreadyVotedForAnotherPeerInTheSameTerm(t *testing.T) {
	n := getNode(nil, nil, nil)
	currentTerm, ok := n.d.store.getInt(currentTermKey)
	n.d.store.setValue(votedForKey, "peer2")
	n.d.store.setInt(voteGrantedInTermKey, currentTerm)

	if !ok {
		t.Fatal("Should have been able to read the current term")
	}

	var actualVoteResponse voteResponse
	n.d.chatter.(*mockChatter).sendVoteResponseStub = func(voteResponse voteResponse) {
		actualVoteResponse = voteResponse
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotRequestForVote, st: n.st, payload: &voteRequest{from: "1", term: (currentTerm)}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if actualVoteResponse.success {
		t.Fatal("Should NOT have got a successful vote response")
	}
}

func Test_AsAFollowerDoesVoteForACandidateWhenItHasPreviouslyVotedForTheSameInTheSameTerm(t *testing.T) {
	n := getNode(nil, nil, nil)
	currentTerm, ok := n.d.store.getInt(currentTermKey)
	n.d.store.setValue(votedForKey, "1")
	n.d.store.setInt(voteGrantedInTermKey, currentTerm)

	if !ok {
		t.Fatal("Should have been able to read the current term")
	}

	var actualVoteResponse voteResponse
	n.d.chatter.(*mockChatter).sendVoteResponseStub = func(voteResponse voteResponse) {
		actualVoteResponse = voteResponse
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotRequestForVote, st: n.st, payload: &voteRequest{from: "1", term: (currentTerm)}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if !actualVoteResponse.success {
		t.Fatal("Should have got a successful vote response")
	}
}

func Test_AsAFollowerDoesVoteForACandidateWhenItHasAlreadyVotedForAnotherPeerForAPreviousTerm(t *testing.T) {
	n := getNode(nil, nil, nil)
	currentTerm, ok := n.d.store.getInt(currentTermKey)
	n.d.store.setValue(votedForKey, "peer2")
	n.d.store.setInt(voteGrantedInTermKey, currentTerm)

	if !ok {
		t.Fatal("Should have been able to read the current term")
	}

	var actualVoteResponse voteResponse
	n.d.chatter.(*mockChatter).sendVoteResponseStub = func(voteResponse voteResponse) {
		actualVoteResponse = voteResponse
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotRequestForVote, st: n.st, payload: &voteRequest{from: "1", term: (currentTerm + 1)}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if !actualVoteResponse.success {
		t.Fatal("Should have got a successful vote response")
	}
}

func Test_AsAFollowerWhenItGetsAnElectionTimeoutAndItHasHeardFromTheLeaderWithinElectionTimeoutItContinuesAsAFollower(t *testing.T) {

	called := 0
	var g getTimerFn = func(d time.Duration) Timer {
		called = called + 1
		return newMockTimer(d)
	}

	n := getNode(g, nil, nil)
	n.d.time = newMockTime((n.lastHeardFromALeader - electionTimeSpan))

	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: GotElectionSignal, st: n.st, payload: nil}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should still have remained as a follower, as it did hear from the leader")
	}

	if called < 2 {
		t.Fatal("Should have invoked the election timer function more than once")
	}
}

func Test_AsAFollowerWhenItGetsAppendEntryItSetsLastHeardFromALeader(t *testing.T) {

	timeNow := int64(150)
	n := getNode(nil, nil, nil)
	n.d.time = newMockTime(timeNow)
	currentTerm, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		t.Fatal("Not able to get current term")
	}
	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: AppendEntry, st: n.st, payload: &appendEntryRequest{term: currentTerm}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should still have remained as a follower, as it did hear from the leader")
	}

	if n.lastHeardFromALeader != timeNow {
		t.Fatal("Should have set lastHeardFromALeader to current time")
	}
}

func Test_AsAFollowerWhenItGetsAppendEntryItRejectsRequestIfItsTermIsLessThanItsOwnCurrentTerm(t *testing.T) {

	timeNow := int64(150)
	n := getNode(nil, nil, nil)
	n.d.time = newMockTime(timeNow)
	var er appendEntryResponse
	n.d.chatter.(*mockChatter).sendAppendEntryResponseStub = func(response appendEntryResponse) {
		er = response
	}
	term := uint64(2)
	n.d.store.setInt(currentTermKey, term)

	directD := n.d.dispatcher.(*directDispatcher)

	evt := event{evtType: AppendEntry, st: n.st, payload: &appendEntryRequest{term: term - 1}}
	directD.dispatch(evt)
	directD.awaitSignal()
	directD.reset()

	if reflect.TypeOf(n.st.stFn) != reflect.TypeOf((*follower)(nil)) {
		t.Fatal("Should still have remained as a follower, as it did hear from the leader")
	}

	if er.success {
		t.Fatal("Should have rejected the append entry request")
	}

	if n.lastHeardFromALeader == timeNow {
		t.Fatal("Should NOT have set lastHeardFromALeader to current time")
	}
}
