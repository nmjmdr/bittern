package raft

type EventType int

const (
	Start EventType = iota
	Stop
	AppendEntry
	GotVote
	GotVoteRejected
	GotRequestForVote
	GotElectionSignal
	GotHeartbeatSignal
)

type event struct {
	evtType EventType
	st      *state
	payload interface{}
}
