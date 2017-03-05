package raft

import (
	"math/rand"
	"time"
)

const ElectionTimeoutMax = 150
const ElectionTimeoutMin = 100

func getRandomizedElectionTimout() time.Duration {
	return time.Duration((rand.Intn(ElectionTimeoutMax-ElectionTimeoutMin) + ElectionTimeoutMin)) * time.Millisecond
}
