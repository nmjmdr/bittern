package raft

import (
	"math/rand"
)

const ElectionTimeoutMax = 150
const ElectionTimeoutMin = 100

func getRandomizedElectionTimout() int64 {
	return int64(rand.Intn(ElectionTimeoutMax-ElectionTimeoutMin) + ElectionTimeoutMin)
}

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func getCurrentTerm(n *node) uint64 {
	term, ok := n.store.GetInt(CurrentTermKey)
	if !ok {
		panic("Not able to to obtain current term")
	}
	return term
}
