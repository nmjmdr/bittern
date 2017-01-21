package raft

type voteRequest struct {
  from string
  term uint64
}
