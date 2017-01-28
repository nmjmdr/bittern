package raft

type EventType int

const (
	Start EventType = iota
	Stop
	AppendEntry
	GotVote
	GotVoteRequestRejected
	GotRequestForVote
	GotElectionSignal
	GotHeartbeatSignal
)

type event struct {
	evtType EventType
	st      *state
	payload interface{}
}
