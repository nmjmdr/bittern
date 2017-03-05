package raft

import (
	"math/rand"
)

const ElectionTimeoutMax = 150
const ElectionTimeoutMin = 100

func getRandomizedElectionTimout() int64 {
	return int64(rand.Intn(ElectionTimeoutMax-ElectionTimeoutMin) + ElectionTimeoutMin)
}
