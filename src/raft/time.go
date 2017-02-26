package raft

import (
  "time"
)

type Time interface {
  unixNano() int64
}

type machineTime struct {
}

func (m *machineTime) unixNano() int64 {
  return time.Now().UnixNano()
}
