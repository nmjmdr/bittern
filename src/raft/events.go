package raft

type EventType int

const (
	StartFollower EventType = iota
	ElectionTimerTimedout
	StartCandidate
	StartLeader
	GotVoteResponse
	StepDown
	GotRequestForVote
	AppendEntries
)

type event struct {
	eventType EventType
	payload   interface{}
}
