package raft

type appendEntryRequest struct {
  from peer
  term uint64
}
