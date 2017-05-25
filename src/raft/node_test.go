package raft

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func createNamedNode(id string) *node {
	n := newNode(id)
	n.dispatcher = newMockDispathcer()
	n.store = newInMemoryStore()
	n.votedFor = newVotedForStore(n.store)
	n.electionExpiryTimer = newMockTimer()
	n.heartbeatTimer = newMockTimer()
	n.log = newMockLog(uint64(0), uint64(0))
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{peer{"1"}}
	})
	delta := int64(20)
	n.time = newMockTime(func() int64 {
		return n.electionTimeout - delta
	})
	n.campaigner = newMockCampaigner()
	n.transport = newMockTransport()
	return n
}

func makeNodeLeader(n *node) {
	n.st.mode = Leader
	n.dispatcher.Dispatch(event{StartLeader, nil})
}

func createNode() *node {
	return createNamedNode("peer0")
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
	currentTerm := getCurrentTerm(n)
	if currentTerm != 0 {
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
	n.electionExpiryTimer.(*mockTimer).startCb = func(t time.Duration) {
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

func Test_when_the_mode_is_candidate_and_election_timer_times_out__and_the_node_has_not_heard_from_leader_it_starts_as_a_candidate_again(t *testing.T) {
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
	previousTerm := getCurrentTerm(n)
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	n.dispatcher.Dispatch(event{StartCandidate, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	var term uint64
	term = getCurrentTerm(n)
	if term != previousTerm+1 {
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
	n.dispatcher.Dispatch(event{StartCandidate, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	votesGot := n.st.votesGot
	if votesGot != holdVotesGot+1 {
		t.Fatal("It should have incremented votesGot by 1")
	}
}

func Test_when_the_mode_is_candidate_and_start_candidate_is_handled_it_restarts_election(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	startTimerCalled := false
	stopTimerCalled := false
	n.electionExpiryTimer.(*mockTimer).startCb = func(t time.Duration) {
		startTimerCalled = true
	}
	n.electionExpiryTimer.(*mockTimer).stopCb = func() {
		stopTimerCalled = true
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	n.dispatcher.Dispatch(event{StartCandidate, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
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
	n.campaigner.(*mockCampaigner).callback = func(node *node) {
		campaignerCalled = true
	}
	n.dispatcher.Dispatch(event{StartCandidate, nil})
	if n.st.mode != Candidate {
		t.Fatal("Should have been a Candidate")
	}
	if !campaignerCalled {
		t.Fatal("It should have called the campaigner")
	}
}

func Test_when_the_mode_is_candidate_and_it_gets_majority_votes_it_should_transition_to_a_leader(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	startLeaderEventGenerated := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartLeader {
			startLeaderEventGenerated = true
		} else {
			n.handleEvent(event)
		}
	}
	peers := []peer{peer{"1"}, peer{"2"}, peer{"3"}}
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return peers
	})
	// start Candidate so that it votes for itself
	n.dispatcher.Dispatch(event{StartCandidate, nil})
	if n.st.votesGot != 1 {
		t.Fatal("A node in candidate mode, should have voted for itself")
	}
	for i := 0; i < len(peers)-1; i++ {
		n.dispatcher.Dispatch(event{GotVoteResponse, &voteResponse{true, 0, peer{}}})
	}
	if !startLeaderEventGenerated {
		t.Fatal("Should have got elected as a leader")
	}
}

func Test_when_the_mode_is_candidate_and_it_gets_a_rejected_vote_it_steps_down(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	steppedDown := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartFollower {
			steppedDown = true
		} else {
			n.handleEvent(event)
		}
	}
	term := getCurrentTerm(n)
	n.dispatcher.Dispatch(event{GotVoteResponse, &voteResponse{false, (term + 1), peer{}}})
	if !steppedDown {
		t.Fatal("Should have stepped down")
	}
}

func Test_when_as_a_follower_node_gets_request_for_vote_it_rejects_it_if_the_request_term_is_less_than_its_own_term(t *testing.T) {
	id := "peer0"
	n := createNamedNode(id)
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var gotVoteResponse voteResponse
	peerRequestingVote := "peer1"
	var voteResponseSentTo peer
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		voteResponseSentTo = sendToPeer
	}
	term := uint64(2)
	n.store.StoreInt(CurrentTermKey, term)
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{peerRequestingVote}, term: term - 1}})

	if voteResponseSentTo.id != peerRequestingVote {
		t.Fatal(fmt.Sprintf("Should have sent the rejection to %s, but got it for: %s", peerRequestingVote, voteResponseSentTo.id))
	}
	if gotVoteResponse.success {
		t.Fatal("Should have rejected the vote request")
	}
	if gotVoteResponse.from.id != id {
		t.Fatal(fmt.Sprintf("Should have got the response from: %s, but got it from: %s", id, gotVoteResponse.from.id))
	}
}

func Test_when_new_term_is_passed_to_voted_for_it_returns_empty_as_candidate_id(t *testing.T) {
	votedFor := newVotedForStore(newInMemoryStore())
	term := uint64(1)
	candidateId := votedFor.Get(term)
	if candidateId != "" {
		t.Fatal("Should have returned an empty candidate id")
	}
}

func Test_when_an_existing_term_is_passed_to_voted_for_it_returns_the_candidate_id_that_was_saved_earlier(t *testing.T) {
	votedFor := newVotedForStore(newInMemoryStore())
	term := uint64(1)
	candidateId := "peer1"
	votedFor.Store(term, candidateId)
	storedCandidateId := votedFor.Get(term)
	if storedCandidateId != candidateId {
		t.Fatal("Should have returned the candidate id that was saved earlier")
	}
}

func Test_when_a_new_term_is_passed_to_voted_for_and_there_exists_a_value_for_different_term_it_returns_an_empty_candidate_id(t *testing.T) {
	votedFor := newVotedForStore(newInMemoryStore())
	term := uint64(1)
	candidateId := "peer1"
	votedFor.Store(term, candidateId)
	newTerm := uint64(2)
	storedCandidateId := votedFor.Get(newTerm)
	if storedCandidateId != "" {
		t.Fatal("Should have returned an empty candidateId")
	}
}

func Test_when_there_exists_a_previous_value_in_voted_and_a_new_voted_for_values_are_stored_it_then_returns_the_new_values(t *testing.T) {
	votedFor := newVotedForStore(newInMemoryStore())
	term := uint64(1)
	candidateId := "peer1"
	votedFor.Store(term, candidateId)
	storedCandidateId := votedFor.Get(term)
	if storedCandidateId != candidateId {
		t.Fatal("Should have returned the previously stored candidateId")
	}

	newTerm := uint64(2)
	newCandidateId := "peer2"
	votedFor.Store(newTerm, newCandidateId)
	storedCandidateId = votedFor.Get(newTerm)
	if storedCandidateId != newCandidateId {
		t.Fatal("Should have returned an newly stored candidateId")
	}
}

func Test_when_a_node_has_already_voted_for_another_peer_in_a_given_term_then_it_rejects_the_request_for_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var voteResponseSentTo peer
	var gotVoteResponse voteResponse
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		voteResponseSentTo = sendToPeer
	}
	term := getCurrentTerm(n)
	n.log.(*mockLog).lastLogTerm = term
	peerVotedForEarlier := "peer1"
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{peerVotedForEarlier}, term: term}})
	if (voteResponseSentTo.id != peerVotedForEarlier) || (!gotVoteResponse.success) {
		t.Fatal(fmt.Sprintf("Should have got a successful vote for %s", peerVotedForEarlier))
	}
	anotherPeerRequestingVote := "peer2"
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{anotherPeerRequestingVote}, term: term}})
	if gotVoteResponse.success {
		t.Fatal(fmt.Sprintf("Should NOT have got a successful vote for %s", anotherPeerRequestingVote))
	}
}

func Test_when_a_node_has_already_voted_for_a_peer_in_a_given_term_and_the_same_peer_requests_for_vote_again_then_it_returns_a_success_response(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var voteResponseSentTo peer
	var gotVoteResponse voteResponse
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		voteResponseSentTo = sendToPeer
	}
	term := getCurrentTerm(n)
	n.log.(*mockLog).lastLogTerm = term
	peerVotedForEarlier := "peer1"
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{peerVotedForEarlier}, term: term}})
	if (voteResponseSentTo.id != peerVotedForEarlier) || (!gotVoteResponse.success) {
		t.Fatal(fmt.Sprintf("Should have got a successful vote for %s", peerVotedForEarlier))
	}
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{peerVotedForEarlier}, term: term}})
	if !gotVoteResponse.success {
		t.Fatal(fmt.Sprintf("Should have got a successful vote for %s", peerVotedForEarlier))
	}
}

func Test_when_a_node_has_already_voted_for_another_peer_in_a_previous_term_and_another_peer_requests_for_vote_in_another_term_it_gets_a_success_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var voteResponseSentTo peer
	var gotVoteResponse voteResponse
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		voteResponseSentTo = sendToPeer
	}
	term := getCurrentTerm(n)
	n.log.(*mockLog).lastLogTerm = term
	peerVotedForEarlier := "peer1"
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{peerVotedForEarlier}, term: term}})
	if (voteResponseSentTo.id != peerVotedForEarlier) || (!gotVoteResponse.success) {
		t.Fatal(fmt.Sprintf("Should have got a successful vote for %s", peerVotedForEarlier))
	}
	anotherPeerRequestingVote := "peer2"
	anotherTerm := term + 1
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{anotherPeerRequestingVote}, term: anotherTerm}})
	if !gotVoteResponse.success {
		t.Fatal(fmt.Sprintf("Should have got a successful vote for %s", anotherPeerRequestingVote))
	}
}

func Test_when_the_nodes_last_log_term_is_greater_than_vote_requests_last_log_term_it_rejects_the_request_for_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var gotVoteResponse voteResponse
	sendVoteResponseInvoked := false
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		sendVoteResponseInvoked = true
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	nodesLastLogTerm := uint64(2)
	n.log.(*mockLog).lastLogTerm = nodesLastLogTerm
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: term, lastLogTerm: (nodesLastLogTerm - 1)}})
	if !sendVoteResponseInvoked {
		t.Fatal("Should have sent a vote response")
	}
	if gotVoteResponse.success {
		t.Fatal("Should have rejected the vote request")
	}
}

func Test_when_the_nodes_log_term_is_same_as_the_vote_requests_last_log_term_but_nodes_log_length_is_greater_it_rejects_the_request_for_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var gotVoteResponse voteResponse
	sendVoteResponseInvoked := false
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		sendVoteResponseInvoked = true
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	nodesLastLogTerm := uint64(2)
	n.log.(*mockLog).lastLogTerm = nodesLastLogTerm
	index := uint64(2)
	n.log.(*mockLog).lastLogIndex = index
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: term, lastLogTerm: (nodesLastLogTerm), lastLogIndex: (index - 1)}})
	if !sendVoteResponseInvoked {
		t.Fatal("Should have sent a vote response")
	}
	if gotVoteResponse.success {
		t.Fatal("Should have rejected the vote request")
	}
}

func Test_when_the_nodes_last_log_term_is_smaller_than_vote_requests_last_log_term_it_accepts_the_request_for_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var gotVoteResponse voteResponse
	sendVoteResponseInvoked := false
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		sendVoteResponseInvoked = true
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	nodesLastLogTerm := uint64(2)
	n.log.(*mockLog).lastLogTerm = nodesLastLogTerm
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: term, lastLogTerm: (nodesLastLogTerm + 1)}})
	if !sendVoteResponseInvoked {
		t.Fatal("Should have sent a vote response")
	}
	if !gotVoteResponse.success {
		t.Fatal("Should have accepted the vote request")
	}
}

func Test_when_the_nodes_last_log_term_is_smaller_than_vote_requests_last_log_term_but_its_last_log_index_is_higher_it_still_accepts_the_request_for_vote(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var gotVoteResponse voteResponse
	sendVoteResponseInvoked := false
	n.transport.(*mockTransport).sendVoteResponseCb = func(sendToPeer peer, voteResponse voteResponse) {
		gotVoteResponse = voteResponse
		sendVoteResponseInvoked = true
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	nodesLastLogTerm := uint64(2)
	n.log.(*mockLog).lastLogTerm = nodesLastLogTerm
	index := uint64(2)
	n.log.(*mockLog).lastLogIndex = index
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: term, lastLogTerm: (nodesLastLogTerm + 1), lastLogIndex: (index - 1)}})
	if !sendVoteResponseInvoked {
		t.Fatal("Should have sent a vote response")
	}
	if !gotVoteResponse.success {
		t.Fatal("Should have accepted the vote request")
	}
}

func Test_when_the_node_is_follower_and_it_gets_a_request_for_vote_with_higher_term_it_sets_its_term_to_the_higher_term(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	higherTerm := term + 1
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: higherTerm}})
	setTerm := getCurrentTerm(n)
	if setTerm != higherTerm {
		t.Fatal("Should have set the term to the higher term that was discovered")
	}
}

func Test_when_the_node_is_candiate_and_it_gets_a_request_for_vote_with_higher_term_it_steps_down(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	verifyIfTheNodeStepsDownIsRaisedWhenHigherTermIsDiscovered(n, t)
}

func Test_when_the_node_is_Leader_and_it_gets_a_request_for_vote_with_higher_term_it_steps_down(t *testing.T) {
	n := createNode()
	n.boot()
	verifyIfTheNodeStepsDownIsRaisedWhenHigherTermIsDiscovered(n, t)
}

func verifyIfTheNodeStepsDownIsRaisedWhenHigherTermIsDiscovered(n *node, t *testing.T) {
	stepDownRaised := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartFollower {
			stepDownRaised = true
		}
		n.handleEvent(event)
	}
	term := uint64(1)
	n.store.StoreInt(CurrentTermKey, term)
	higherTerm := term + 1
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{GotRequestForVote, &voteRequest{from: peer{"some-peer"}, term: higherTerm}})
	if !stepDownRaised {
		t.Fatal("Should have raised step down event")
	}
}

func Test_when_the_node_receives_an_append_entries_with_a_term_less_than_its_own_it_rejects_it(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	sendAppendEntriesResponseCalled := false
	var gotAppendEntriesResponse appendEntriesResponse
	n.transport.(*mockTransport).sendAppendEntriesResponseCb = func(sendToPeer peer, ar appendEntriesResponse) {
		sendAppendEntriesResponseCalled = true
		gotAppendEntriesResponse = ar
	}
	term := uint64(2)
	n.store.StoreInt(CurrentTermKey, term)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: (term - 1)}})
	if !sendAppendEntriesResponseCalled {
		t.Fatal("Should have called send append entry response")
	}
	if gotAppendEntriesResponse.success {
		t.Fatal("Should have rejected the append entry response")
	}
	if gotAppendEntriesResponse.term != term {
		t.Fatal("Should have set the response term to rejecting node's term")
	}
}

// append entries, rule 2. Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm (§5.3)
func Test_when_the_node_receives_an_append_entries_and_the_entry_at_prev_log_index_does_not_match_the_prev_log_term_on_term_it_rejects_it(t *testing.T) {
	n := createNode()
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	sendAppendEntriesResponseCalled := false
	var gotAppendEntriesResponse appendEntriesResponse
	n.transport.(*mockTransport).sendAppendEntriesResponseCb = func(sendToPeer peer, ar appendEntriesResponse) {
		sendAppendEntriesResponseCalled = true
		gotAppendEntriesResponse = ar
	}
	term := uint64(2)
	n.store.StoreInt(CurrentTermKey, term)
	prevLogIndex := uint64(10)
	var indexPassed uint64
	entryAtPrevLogIndex := entry{term: (term - 1)}
	n.log.(*mockLog).entryAtCb = func(index uint64) (entry, bool) {
		indexPassed = index
		return entryAtPrevLogIndex, true
	}
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex}})
	if !sendAppendEntriesResponseCalled {
		t.Fatal("Should have called send append entry response")
	}
	if indexPassed != prevLogIndex {
		t.Fatal("Should have passed prevLogIndex: %d to check for entry at log", prevLogIndex)
	}
	if gotAppendEntriesResponse.success {
		t.Fatal("Should have rejected the append entry response")
	}

	n.log.(*mockLog).entryAtCb = func(index uint64) (entry, bool) {
		indexPassed = index
		// no entry was found
		return entryAtPrevLogIndex, false
	}
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex}})
	if !sendAppendEntriesResponseCalled {
		t.Fatal("Should have called send append entry response")
	}
	if indexPassed != prevLogIndex {
		t.Fatal("Should have passed prevLogIndex: %d to check for entry at log", prevLogIndex)
	}
	if gotAppendEntriesResponse.success {
		t.Fatal("Should have rejected the append entry response as the test would have returned as no entry found at prevLogIndex")
	}
}

func Test_when_there_is_a_mismatched_entry_in_log_append_entries_removes_it_and_those_that_follow_it(t *testing.T) {
	n := createNode()
	n.boot()
	entryTerm := uint64(2)
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{term: entryTerm}, true
	}
	deleteFromInvoked := false
	deleteFromIndex := uint64(0)
	n.log.(*mockLog).deleteFromCb = func(index uint64) {
		deleteFromIndex = index
		deleteFromInvoked = true
	}
	n.log.(*mockLog).addAtCb = func(logIndex uint64, e entry) {
	}
	index := uint64(10)
	n.appendToLog(&appendEntriesRequest{prevLogIndex: index, entries: []entry{entry{term: (entryTerm + 1)}}})
	if !deleteFromInvoked {
		t.Fatal("Delete from index was NOT invoked")
	}
	if deleteFromIndex != index {
		t.Fatal("Should have invoked delete from index starting at prevLogIndex")
	}
}

func Test_when_the_nodes_log_is_empty_and_all_other_conditions_met_it_accepts_log_entries(t *testing.T) {
	n := createNode()
	n.boot()
	addAtCalled := false
	var entryAdded entry
	n.log.(*mockLog).addAtCb = func(logIndex uint64, e entry) {
		addAtCalled = true
		entryAdded = e
	}
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	term := getCurrentTerm(n)
	prevLogIndex := uint64(0)
	command := "command"
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term, command: command}}}})
	if !addAtCalled {
		t.Fatal("Should have called log.AddAt")
	}
	if entryAdded.command != command {
		t.Fatal("Should have got the entry that was passed as part of append entry request")
	}
}

func Test_when_the_nodes_is_a_candidate_and_it_accepts_log_entries_from_the_new_leader_it_steps_down(t *testing.T) {
	n := createNode()
	n.boot()
	n.st.mode = Candidate
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	steppedDown := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartFollower {
			steppedDown = true
		} else {
			n.handleEvent(event)
		}
	}
	term := getCurrentTerm(n)
	prevLogIndex := uint64(0)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}}})
	if !steppedDown {
		t.Fatal("Should have stepped down")
	}
}

func Test_when_the_nodes_is_a_leader_and_it_gets_append_entries_from_another_leader_with_higher_term_it_steps_down(t *testing.T) {
	n := createNode()
	n.boot()
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	steppedDown := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartFollower {
			steppedDown = true
		} else {
			n.handleEvent(event)
		}
	}
	makeNodeLeader(n)
	term := getCurrentTerm(n)
	higherTerm := term + 1
	prevLogIndex := uint64(0)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: higherTerm, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}}})
	if !steppedDown {
		t.Fatal("Should have called step down")
	}
}

func Test_when_the_nodes_is_a_follower_and_it_gets_append_entries_from_a_leader_with_higher_term_it_sets_its_own_term_to_higher_term(t *testing.T) {
	n := createNode()
	n.boot()
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}

	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	term := getCurrentTerm(n)
	higherTerm := term + 1
	prevLogIndex := uint64(0)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: higherTerm, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}}})
	setTerm := getCurrentTerm(n)
	if setTerm != higherTerm {
		t.Fatal("Should have set to higher term")
	}
}

func Test_when_node_receieves_and_accepts_log_entries_from_the_new_leader_and_sets_the_commit_index_to_lower_of_leaders_commit_index_and_nodes_last_log_index(t *testing.T) {
	n := createNode()
	n.boot()
	n.log.(*mockLog).addAtCb = func(logIndex uint64, e entry) {
	}
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	term := getCurrentTerm(n)
	prevLogIndex := uint64(0)
	leaderCommit := uint64(2)
	n.log.(*mockLog).lastLogIndex = leaderCommit + 1
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}, leaderCommit: leaderCommit}})
	if n.st.commitIndex != leaderCommit {
		t.Fatal("Should have set the commit index to leader's commit")
	}
	n.log.(*mockLog).lastLogIndex = leaderCommit - 1
	n.st.commitIndex = 0
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}, leaderCommit: leaderCommit}})
	if n.st.commitIndex != n.log.(*mockLog).lastLogIndex {
		t.Fatal("Should have set the commit index to last log index")
	}
}

func Test_when_the_nodes_accepts_log_entries_from_the_new_leader_it_sets_last_heard_from_a_leader(t *testing.T) {
	n := createNode()
	n.boot()
	n.log.(*mockLog).addAtCb = func(logIndex uint64, e entry) {
	}
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	timeNow := int64(100)
	n.time.(*mockTime).cb = func() int64 {
		return timeNow
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	term := getCurrentTerm(n)
	prevLogIndex := uint64(0)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}}})
	if n.st.lastHeardFromALeader != timeNow {
		t.Fatal("Should have set last heard from a leader")
	}
}

func Test_when_the_node_starts_as_leader_it_sends_initial_heartbeat_to_all_nodes(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{peer{"peer1"}, peer{"peer2"}}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	sendAppendEntriesRequestCalled := false
	var sentHeartbeatToPeers []peer
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(peer peer, ar appendEntriesRequest) {
		sendAppendEntriesRequestCalled = true
		sentHeartbeatToPeers = append(sentHeartbeatToPeers, peer)
	}
	makeNodeLeader(n)
	if !sendAppendEntriesRequestCalled && len(sentHeartbeatToPeers) != len(n.whoArePeers.All()) {
		t.Fatal("Should have sent append entires to all peers")
	}
}

func Test_when_the_node_starts_as_leader_it_starts_heartbeat_timer(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{}
	})
	heartbeatTimerStartCalled := false
	n.heartbeatTimer.(*mockTimer).startCb = func(t time.Duration) {
		heartbeatTimerStartCalled = true
	}
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(p peer, ar appendEntriesRequest) {
	}
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	makeNodeLeader(n)
	if !heartbeatTimerStartCalled {
		t.Fatal("Should have called heartbear timer start")
	}
}

func Test_when_the_leader_receives_heartbeat_timer_timedout_and_it_has_sent_append_entries_within_time_between_heartbeats_it_does_not_send_heartbeat(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	sendAppendEntriesRequestCalled := false
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(p peer, ar appendEntriesRequest) {
		sendAppendEntriesRequestCalled = true
	}
	timeNow := int64(100)
	n.time = newMockTime(func() int64 {
		return timeNow
	})
	n.st.lastSentAppendEntriesAt = (timeNow - timeBetweenHeartbeats + 10)
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{HeartbeatTimerTimedout, nil})
	if sendAppendEntriesRequestCalled {
		t.Fatal("Should NOT have sent heartbeat")
	}
}

func Test_when_the_leader_receives_heartbeat_timer_timedout_and_it_has_NOT_sent_append_entries_within_time_between_heartbeats_it_sends_heartbeat(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{peer{"peer1"}}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	startTimerCalled := false
	n.heartbeatTimer.(*mockTimer).startCb = func(t time.Duration) {
		startTimerCalled = true
	}
	sendAppendEntriesRequestCalled := false
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(p peer, ar appendEntriesRequest) {
		sendAppendEntriesRequestCalled = true
	}
	timeNow := int64(100)
	n.time = newMockTime(func() int64 {
		return timeNow
	})
	n.st.lastSentAppendEntriesAt = (timeNow - timeBetweenHeartbeats - 10)
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{HeartbeatTimerTimedout, nil})
	if !sendAppendEntriesRequestCalled {
		t.Fatal("Should have sent heartbeat")
	}
}

func Test_when_the_leader_receives_heartbeat_timer_timedout_it_restarts_the_heartbeat_timer(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	startTimerCalled := false
	n.heartbeatTimer.(*mockTimer).startCb = func(t time.Duration) {
		startTimerCalled = true
	}
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{HeartbeatTimerTimedout, nil})
	if !startTimerCalled {
		t.Fatal("Should have called start timer for heartbeat timer")
	}
}

func Test_when_the_leader_receives_append_entries_response_it_steps_down_if_it_discovers_a_higher_term(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{}
	})
	n.boot()
	steppedDown := false
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		if event.eventType == StartFollower {
			steppedDown = true
		} else {
			n.handleEvent(event)
		}
	}
	term := uint64(2)
	n.store.StoreInt(CurrentTermKey, term)
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{GotAppendEntriesResponse, &appendEntriesResponse{success: false, term: (term + 1)}})
	if !steppedDown {
		t.Fatal("Should have raised the step down event")
	}
}

func Test_when_the_nodes_accepts_log_entries_from_a_leader_it_sends_a_successful_append_entries_response(t *testing.T) {
	n := createNode()
	n.boot()
	n.log.(*mockLog).addAtCb = func(logIndex uint64, e entry) {
	}
	n.log.(*mockLog).entryAtCb = func(logIndex uint64) (entry, bool) {
		return entry{}, false
	}
	timeNow := int64(100)
	n.time.(*mockTime).cb = func() int64 {
		return timeNow
	}
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	var response appendEntriesResponse
	sentAppendEntriesResponse := false
	n.transport.(*mockTransport).sendAppendEntriesResponseCb = func(to peer, ar appendEntriesResponse) {
		sentAppendEntriesResponse = true
		response = ar
	}
	term := getCurrentTerm(n)
	prevLogIndex := uint64(0)
	n.dispatcher.Dispatch(event{AppendEntries, &appendEntriesRequest{from: peer{"peer1"}, term: term, prevLogTerm: term, prevLogIndex: prevLogIndex, entries: []entry{entry{term: term}}}})
	if !sentAppendEntriesResponse || !response.success {
		t.Fatal("Should have sent a successful append entries response")
	}
}

func Test_when_the_leader_receives_a_command_it_appends_it_to_its_log(t *testing.T) {
	n := createNamedNode("peer0")
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}

	appendToLogInvoked := false
	term := uint64(2)
	command := "command"
	n.store.StoreInt(CurrentTermKey, term)
	var entryAppended entry
	(n.log.(*mockLog)).appendCb = func(e entry) {
		appendToLogInvoked = true
		entryAppended = e
	}
	makeNodeLeader(n)
	n.dispatcher.Dispatch(event{GotCommand, command})
	if !appendToLogInvoked {
		t.Fatal("Should have invoked append to log")
	}
}

func Test_when_the_leader_receives_a_command_it_sends_append_entries_for_replication(t *testing.T) {
	n := createNamedNode("peer0")
	toId := "peer1"
	toPeer := peer{id: toId}
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{toPeer}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}

	term := uint64(2)
	command := "command"

	n.log.(*mockLog).getCb = func(startIndex uint64) []entry {
		return []entry{entry{term: term, command: command}}
	}

	n.store.StoreInt(CurrentTermKey, term)
	makeNodeLeader(n)
	appendEntriesSent := false
	commandReceived := ""
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(sentToPeer peer, ar appendEntriesRequest) {
		appendEntriesSent = true
		if ar.entries != nil && len(ar.entries) >= 1 {
			commandReceived = ar.entries[0].command
		}
	}

	n.dispatcher.Dispatch(event{GotCommand, command})
	if !appendEntriesSent || commandReceived != command {
		t.Fatal("Should have sent the command  to other peer")
	}
}

func Test_when_the_leader_starts_it_resets_the_next_index_and_match_index_maps(t *testing.T) {
	n := createNamedNode("peer0")
	peerId := "peer1"
	other := peer{id: peerId}
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{other}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	lastLogIndex := uint64(10)
	n.log = newMockLog(lastLogIndex, 2)
	makeNodeLeader(n)
	if n.st.nextIndex[other.id] != lastLogIndex+1 {
		t.Fatal("Should re-initialized next-index to last log index + 1")
	}
	if n.st.matchIndex[other.id] != 0 {
		t.Fatal("Should re-initialized match-index to 0")
	}
}

func Test_when_the_leader_receives_a_failed_append_entries_response_when_the_peer_log_does_not_match_it_decrements_the_next_index_for_the_peer(t *testing.T) {
	n := createNamedNode("peer0")
	peerId := "peer1"
	other := peer{id: peerId}
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{other}
	})
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	lastLogIndex := uint64(10)
	n.log = newMockLog(lastLogIndex, 2)
	makeNodeLeader(n)
	term := getCurrentTerm(n)
	nextIndexForPeer := uint64(10)
	n.st.nextIndex[peerId] = nextIndexForPeer
	n.dispatcher.Dispatch(event{GotAppendEntriesResponse, &appendEntriesResponse{success: false, term: term, from: peerId}})
	if n.st.nextIndex[peerId] != (nextIndexForPeer - 1) {
		t.Fatal("Should have decremented the next index for the peer")
	}
}

func Test_leader_sends_all_commands_in_range_from_next_index_to_last_log_index_to_peers(t *testing.T) {
	n := createNamedNode("peer0")
	peerId := "peer1"
	other := peer{id: peerId}
	n.whoArePeers = newMockWhoArePeers(func() []peer {
		return []peer{other}
	})
	var entries []entry
	n.transport.(*mockTransport).sendAppendEntriesRequestCb = func(p peer, ar appendEntriesRequest) {
		entries = ar.entries
	}
	n.boot()
	n.dispatcher.(*mockDispatcher).callback = func(event event) {
		n.handleEvent(event)
	}
	n.log = newInmemoryLog()

	makeNodeLeader(n)
	// now the leader receives three commands
	command := "command"
	numberOfCommands := 3
	var commands []string
	for i := 0; i < numberOfCommands; i++ {
		commands = append(commands, command + strconv.Itoa(i))
		n.dispatcher.Dispatch(event{GotCommand, commands[i] })
	}
	if len(entries) != numberOfCommands {
		t.Fatal("Should have sent all the commands to peer")
	}
	var index = 0
	for _,e := range entries {
		if e.command != commands[index] {
			t.Fatal("Should have recieved the command")
		}
		index++
	}
}
