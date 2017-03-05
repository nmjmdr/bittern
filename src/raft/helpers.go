package raft

import (
  "time"
  "math/rand"
)

const ElectionTimeoutMax = 150
const ElectionTimeoutMin = 100

func getRandomizedElectionTimout() time.Duration {
  return time.Duration((rand.Intn(ElectionTimeoutMax - ElectionTimeoutMin) + ElectionTimeoutMin)) * time.Millisecond
}
