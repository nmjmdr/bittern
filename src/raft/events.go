package raft

type EventType int

const (
	StartFollower EventType = iota
	ElectionTimerTimedout
	StartCandidate
	StartLeader
	GotVote
)

type event struct {
	eventType EventType
	payload   interface{}
}
