package raft

type Dispatcher interface {
  Dispatch(event event)
}
