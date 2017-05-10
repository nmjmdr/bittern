package raft

type EventType int

const (
	StartFollower EventType = iota
	ElectionTimerTimedout
	StartCandidate
	StartLeader
	GotVoteResponse
	GotRequestForVote
	AppendEntries
	HeartbeatTimerTimedout
	GotAppendEntriesResponse
	GotCommand
)

type event struct {
	eventType EventType
	payload   interface{}
}
