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
	n.campaigner = newMockCampaigner(nil)
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
	n.time = newMockTime(func() int64 {
		return n.electionTimeout + delta
	})
	startCandidateEventDispatched := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == ElectionTimerTimedout {
			n.handleEvent(event)
		} else if event.eventType == StartCandidate {
			startCandidateEventDispatched = true
		} else {
			t.Fatal("Not expecting %d event to be raised", event.eventType)
		}
	}
	n.dispatcher.Dispatch(event{ElectionTimerTimedout, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	if !startCandidateEventDispatched {
		t.Fatal("Should have dispatched StartCandidate event")
	}
}


func Test_when_the_mode_is_candidate_and_election_timer_timesout__and_has_not_heard_from_leader_it_starts_a_candidate_again(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	delta := int64(20)
	n.time = newMockTime(func() int64 {
		return n.electionTimeout + delta
	})
	startCandidateEventDispatched := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == ElectionTimerTimedout {
			n.handleEvent(event)
		} else if event.eventType == StartCandidate {
			startCandidateEventDispatched = true
		} else {
			t.Fatal("Not expecting %d event to be raised", event.eventType)
		}
	}
	n.dispatcher.Dispatch(event{ElectionTimerTimedout, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	if !startCandidateEventDispatched {
		t.Fatal("Should have dispatched StartCandidate event")
	}
}

func Test_when_the_mode_is_candidate_and_start_candidate_is_handled_it_increments_the_current_term(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	previousTerm, ok := n.store.GetInt(CurrentTermKey)
	if !ok {
		t.Fatal("Cannot obtain the current term")
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
			n.handleEvent(event)
	}
	n.dispatcher.Dispatch(event{StartCandidate,nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	var term uint64
	term,ok = n.store.GetInt(CurrentTermKey)
	if !ok {
		t.Fatal("Cannot obtain the current term")
	}
	if term != previousTerm + 1 {
		t.Fatal("It should have incremented the term by 1")
	}
}

func Test_when_the_mode_is_candidate_and_start_candidate_is_handled_it_votes_for_itself(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	holdVotesGot := n.st.votesGot
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
			n.handleEvent(event)
	}
	n.dispatcher.Dispatch(event{StartCandidate,nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	votesGot := n.st.votesGot
	if votesGot != holdVotesGot + 1 {
		t.Fatal("It should have incremented votesGot by 1")
	}
}

func Test_when_the_mode_is_candidate_and_start_candidate_is_handled_it_restarts_election(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
			n.handleEvent(event)
	}
	n.dispatcher.Dispatch(event{StartCandidate,nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	startTimerCalled := false
	stopTimerCalled := true
	n.electionExpiryTimer = newMockElectionTimeoutTimer(func(t time.Duration){
		startTimerCalled = true
	},func(){
		stopTimerCalled = true
	})
	if !stopTimerCalled {
		t.Fatal("It should have called stop election timer")
	}
	if !startTimerCalled {
		t.Fatal("It should have called start election timer")
	}
}


func Test_when_the_mode_is_candidate_and_start_candidate_is_handled_it_starts_the_campaign(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
			n.handleEvent(event)
	}
	campaignerCalled := false
	n.campaigner = newMockCampaigner(func(node *node){
		campaignerCalled = true
	})
	n.dispatcher.Dispatch(event{StartCandidate,nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	if !campaignerCalled {
		t.Fatal("It should have called the campaigner")
	}
}
