package raft

// TO DO:
// If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)
import (
	"fmt"
	"math/rand"
	"time"
)

const CurrentTermKey = "current-term"
const timeBetweenHeartbeats = 50

type node struct {
	st              *state
	electionTimeout int64
	id              string

	dispatcher          Dispatcher
	store               Store
	electionExpiryTimer Timer
	time                Time
	campaigner          Campaigner
	whoArePeers         WhoArePeers
	transport           Transport
	votedFor            VotedFor
	log                 Log
	heartbeatTimer      Timer
}

func newNode(id string) *node {
	rand.Seed(time.Now().Unix())
	n := new(node)
	n.id = id
	return n
}

func (n *node) boot() {
	n.st = new(state)
	n.store.StoreInt(CurrentTermKey, 0)
	n.st.mode = Follower
	n.st.commitIndex = 0
	n.st.lastApplied = 0

	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) handleEvent(event event) {
	switch event.eventType {
	case StartFollower:
		n.startFollower(event)
	case ElectionTimerTimedout:
		n.electionTimerTimeout(event)
	case StartCandidate:
		n.startCandidate(event)
	case GotVoteResponse:
		n.gotVote(event)
	case GotRequestForVote:
		n.gotRequestForVote(event)
	case AppendEntries:
		n.appendEntries(event)
	case StartLeader:
		n.startLeader(event)
	case HeartbeatTimerTimedout:
		n.heartbeatTimerTimedout(event)
	case GotAppendEntriesResponse:
		n.gotAppendEntriesResponse(event)
	case GotCommand:
		n.gotCommand(event)
	default:
		panic(fmt.Sprintf("Unknown event: %d passed to handleEvent", event.eventType))
	}
}

func (n *node) startElectionTimer() {
	n.electionTimeout = getRandomizedElectionTimout()
	n.electionExpiryTimer.Start(time.Duration(n.electionTimeout) * time.Millisecond)
}

func (n *node) restartElectionTimer() {
	n.electionExpiryTimer.Stop()
	n.startElectionTimer()
}

func (n *node) startFollower(evt event) {
	if n.st.mode != Follower {
		panic("Mode is not set to follower in startFollower")
	}
	n.startElectionTimer()
}

func (n *node) hasHeardFromALeader() bool {
	return (n.time.UnixNow() - n.st.lastHeardFromALeader) < n.electionTimeout
}

func (n *node) electionTimerTimeout(evt event) {
	if n.hasHeardFromALeader() {
		return
	}
	if n.st.mode == Follower {
		n.st.mode = Candidate
		n.dispatcher.Dispatch(event{StartCandidate, nil})
	} else if n.st.mode == Candidate {
		// think about back off from: ARC Heidi Howard - OCAML raft
		n.st.mode = Candidate
		n.dispatcher.Dispatch(event{StartCandidate, nil})
	} else {
		// The following will have to be changed to check if the leader
		// is still able to reach majority of the nodes within an election time out
		panic("Received election timer timedout while being a leader")
	}
}

func (n *node) startCandidate(event event) {
	if n.st.mode != Candidate {
		panic("Mode is not set to candidate in startCandidate")
	}
	term := getCurrentTerm(n)
	term = term + 1
	n.store.StoreInt(CurrentTermKey, term)
	n.st.votesGot = n.st.votesGot + 1
	n.restartElectionTimer()
	n.campaigner.Campaign(n)
}

func (n *node) didIGetMajority() bool {
	peers := n.whoArePeers.All()
	return n.st.votesGot >= (len(peers)/2 + 1)
}

func (n *node) gotVote(evt event) {
	voteResponse := evt.payload.(*voteResponse)
	if n.isHigherTerm(voteResponse.term) {
		n.handleHigherTermReceived(voteResponse.term)

	} else if voteResponse.success {
		n.handleSuccessfulVoteResponse(voteResponse)
	}
}

func (n *node) handleSuccessfulVoteResponse(*voteResponse) {
	n.st.votesGot = n.st.votesGot + 1
	if n.didIGetMajority() {
		n.st.mode = Leader
		n.dispatcher.Dispatch(event{StartLeader, nil})
	}
}

func (n *node) haveIAlreadyVotedForAnotherPeerForTheTerm(term uint64, peerAskingForVote string) bool {
	//NOTE Ref: notes 3.
	candidateVotedFor := n.votedFor.Get(term)
	if candidateVotedFor == "" || candidateVotedFor == peerAskingForVote {
		return false
	}
	return true
}

func (n *node) isCandidatesLogUptoDate(voteRequest *voteRequest) bool {
	// NOTE: Ref Notes 1.
	if voteRequest.lastLogTerm > n.log.LastTerm() {
		return true
	} else if (voteRequest.lastLogTerm == n.log.LastTerm()) && (voteRequest.lastLogIndex >= n.log.LastIndex()) {
		return true
	}
	return false
}

func (n *node) isHigherTerm(termGot uint64) bool {
	term := getCurrentTerm(n)
	return termGot > term
}

func (n *node) handleHigherTermReceived(termGot uint64) {
	n.store.StoreInt(CurrentTermKey, termGot)
	if n.st.mode != Follower {
		n.st.mode = Follower
		n.dispatcher.Dispatch(event{StartFollower, nil})
	}
}

func (n *node) decideVoteResponse(voteRequest *voteRequest, term uint64) voteResponse {
	var response voteResponse
	if n.haveIAlreadyVotedForAnotherPeerForTheTerm(voteRequest.term, voteRequest.from.id) {
		response = voteResponse{false, term, peer{n.id}}
	} else if !n.isCandidatesLogUptoDate(voteRequest) {
		response = voteResponse{false, term, peer{n.id}}
	} else {
		response = voteResponse{true, term, peer{n.id}}
	}
	return response
}

func (n *node) gotRequestForVote(evt event) {
	voteRequest := evt.payload.(*voteRequest)
	term := getCurrentTerm(n)
	if voteRequest.term < term {
		n.transport.SendVoteResponse(voteRequest.from, voteResponse{false, term, peer{n.id}})
		return
	}
	response := n.decideVoteResponse(voteRequest, term)
	if response.success {
		n.votedFor.Store(voteRequest.term, voteRequest.from.id)
	}
	n.transport.SendVoteResponse(voteRequest.from, response)
	if n.isHigherTerm(voteRequest.term) {
		n.handleHigherTermReceived(voteRequest.term)
	}
}

func (n *node) appendToLog(request *appendEntriesRequest) {
	logIndex := request.prevLogIndex
	for _, appendEntry := range request.entries {
		logEntry, ok := n.log.EntryAt(logIndex)
		if ok && logEntry.term != appendEntry.term {
			n.log.DeleteFrom(logIndex)
		}
		n.log.AddAt(logIndex, appendEntry)
		logIndex = logIndex + 1
	}
}

func (n *node) stepDown(term uint64) {
	n.st.mode = Follower
	n.store.StoreInt(CurrentTermKey, term)
	n.dispatcher.Dispatch(event{StartFollower, nil})
}

func (n *node) appendEntries(evt event) {
	request := evt.payload.(*appendEntriesRequest)
	term := getCurrentTerm(n)
	if request.term < term {
		n.transport.SendAppendEntriesResponse(request.from, appendEntriesResponse{false, term})
		return
	}
	// we heard from the leader here: all the other checks below are for log matching, think about this later!!!
	n.st.lastHeardFromALeader = n.time.UnixNow()
	n.ifOkAppendToLog(request, term)
	// Important note: Reference 2
	if n.isHigherTerm(request.term) {
		n.handleHigherTermReceived(request.term)
	} else if n.st.mode == Candidate {
		n.dispatcher.Dispatch(event{StartFollower, nil})
	}
}

func (n *node) ifOkAppendToLog(request *appendEntriesRequest, term uint64) {
	entry, ok := n.log.EntryAt(request.prevLogIndex)
	if ok && entry.term != request.prevLogTerm {
		n.transport.SendAppendEntriesResponse(request.from, appendEntriesResponse{false, term})
		return
	} else if !ok && request.prevLogIndex != 0 {
		// request.prevLogIndex == 0 ==> implies that the leader, when sending the first log entry at index 1 sends prevLogIndex = 0
		n.transport.SendAppendEntriesResponse(request.from, appendEntriesResponse{false, term})
		return
	}
	n.appendToLog(request)
	if request.leaderCommit > n.st.commitIndex {
		n.st.commitIndex = min(request.leaderCommit, n.log.LastIndex())
	}
	n.transport.SendAppendEntriesResponse(request.from, appendEntriesResponse{true, term})
}

func (n *node) sendAppendEntries() {
	term := getCurrentTerm(n)
	peers := n.whoArePeers.All()
	for _, p := range peers {
		n.transport.SendAppendEntriesRequest(p,
			appendEntriesRequest{from: peer{n.id}, term: term, prevLogTerm: n.log.LastTerm(),
				prevLogIndex: n.log.LastIndex(), entries: nil, leaderCommit: n.st.commitIndex})
		n.st.lastSentAppendEntriesAt = n.time.UnixNow()
	}
}

func (n *node) startLeader(evt event) {
	if n.st.mode != Leader {
		panic("startLeader invoked when mode is not set as leader")
	}
	n.sendAppendEntries()
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) heartbeatTimerTimedout(evt event) {
	if n.st.mode != Leader {
		panic("Received heartbeat timer timedout when mode is not set as leader")
	}

	if (n.time.UnixNow() - n.st.lastSentAppendEntriesAt) > timeBetweenHeartbeats {
		n.sendAppendEntries()
	}
	n.heartbeatTimer.Start(time.Duration(timeBetweenHeartbeats) * time.Millisecond)
}

func (n *node) gotCommand(evt event) {
	if n.st.mode != Leader {
		// not a leader, redirect to leader later
		panic("Not a leader, received command. Method - yet to be implemented")
		return
	}
	command := evt.payload.(string)
	currentTerm := getCurrentTerm(n)
	entry := entry{term: currentTerm, command: command}
	n.log.Append(entry)
}

// might have to change this approach: to handle replication response:
//https://groups.google.com/d/topic/raft-dev/M4BraEWG3TU/discussion
/* Work in progress:
Now given that my model of the node is driven by an event-loop and can handle only one event at a time, I can follow the below approach to perform log replication:

1. Append the entry to its log
2. Try and replicate the command by issuing append entries to all peers in parallel (taking into account the match-index and next-index for different peers)
3. Wait for step 2 to come back with success (replicated across majority of peers. What happens when it cannot? Does the leader just hang?)
4. Apply the command
5. Return the response to the client.
Is this ok?
*/
func (n *node) gotAppendEntriesResponse(evt event) {
	appendEntriesResponse := evt.payload.(*appendEntriesResponse)
	if appendEntriesResponse.success && n.st.mode == Leader {
		// process success response here
	} else if n.isHigherTerm(appendEntriesResponse.term) {
		n.handleHigherTermReceived(appendEntriesResponse.term)
	}
	// Look at this for understanding of commiting in the prescence of network partition: https://thesecretlivesofdata.com/raft/
	// Especially: if a leader cannot commit to majority of the nodes, it should stay uncommitted
	// Reference: point 5> in notes, the following rule ensures it:
	// If there exists an N such that N > commitIndex, a majority of matchIndex[i] >= N and log[N].term == currentTerm:
	// then set commitIndex = N
	// Only start from the current commitIndex and then start checking the next one
	// If a leader is not able to send heartbeat or append entry to majority of the nodes within an election timeout
	// then it has to step down
}
