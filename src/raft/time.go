package raft

import (
	"time"
)

type Time interface {
	UnixNow() int64
}

type machineTime struct {
}

func (m *machineTime) UnixNow() int64 {
	return time.Now().Unix()
}
