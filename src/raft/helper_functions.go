package raft

import (
	"math/rand"
	"time"
)

// consult paper to set appropriate times later
const electionTimeoutMaxMs = 300
const electionTimeoutMinMs = 150

func getRandomElectionTimeout() time.Duration {
	n := rand.Intn(electionTimeoutMaxMs-electionTimeoutMinMs) + electionTimeoutMinMs
	return time.Duration(n) * time.Millisecond
}

func getCurrentTerm(store store) uint64 {
	currentTerm, ok := store.getInt(currentTermKey)
	if !ok {
		panic("could not obtain current term as a candiate")
	}
	return currentTerm
}
