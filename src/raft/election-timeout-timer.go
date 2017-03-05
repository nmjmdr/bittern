package raft

import (
  "time"
)

type ElectionTimeoutTimer interface {
	Start(t time.Duration)
}
