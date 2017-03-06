package raft

import (
	"testing"
	"time"
)

func createNode() *node {
	n := newNode()
	n.dispatcher = newMockDispathcer(nil)
	n.store = newInMemoryStore()
	n.electionExpiryTimer = newMockElectionTimeoutTimer(nil, nil)
	delta := int64(20)
	n.time = newMockTime(func() int64 {
		return n.electionTimeout - delta
	})
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
	n.electionExpiryTimer.(*mockElectionTimeoutTimer).startCb = func(t time.Duration) {
		timerStarted = true
	}
	n.boot()
	if !timerStarted {
		t.Fatal("Election timeout timer not started")
	}
}

func Test_when_the_mode_is_follower_and_election_timer_timesout__and_has_not_heard_from_leader_it_transitions_to_a_candidate(t *testing.T) {
	n := createNode()
	n.boot()
	delta := int64(20)
	n.time = newMockTime(func ()int64{
		return n.electionTimeout + delta
	})
	startCandidateEventDispatched := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == ElectionTimerTimedout {
			n.handleEvent(event)
		} else if event.eventType == StartCandidate {
			startCandidateEventDispatched = true
		} else {
			t.Fatal("Not expecting %d event to be raised",event.eventType)
		}
	}
	n.dispatcher.Dispatch(event{ElectionTimerTimedout,nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	if !startCandidateEventDispatched {
		t.Fatal("Shoudl have dispatched StartCandidate event")
	}
}
