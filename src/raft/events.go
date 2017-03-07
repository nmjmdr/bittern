package raft

type EventType int

const (
	StartFollower EventType = iota
	ElectionTimerTimedout
	StartCandidate
	StartLeader
	GotVoteResponse
	GotRejectedVote
	StepDown
)

type event struct {
	eventType EventType
	payload   interface{}
}
