package raft

type EventType int

const (
	StartFollower EventType = iota
	ElectionTimerTimedout
	StartCandidate
)

type event struct {
	eventType EventType
	payload   interface{}
}
