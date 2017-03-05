package raft

type EventType int

const (
	StartFollower EventType = iota
  ElectionTimerTimedout
)

type event struct {
	eventType EventType
	payload   interface{}
}
