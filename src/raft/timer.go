package raft

import (
	"time"
)

type Timer interface {
	Start(t time.Duration)
	Stop()
}
