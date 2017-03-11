package raft

type appendEntryResponse struct {
  success bool
  term uint64
}
