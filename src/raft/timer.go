package raft

import (
  "time"
)

type Timer interface {
  Start(duration time.Duration)
}
