package raft

import (
	"time"
)

var getMockTimerFn getTimerFn = func(d time.Duration) Timer {
	return newMockTimer(d)
}

func getNode(mockTimerFn getTimerFn, store store, peers []peer) *node {
	if store == nil {
		store = newInMemorystore()
	}
	d := newDirectDispatcher()
	if mockTimerFn == nil {
		mockTimerFn = getMockTimerFn
	}
	if peers == nil {
		peers = []peer{}
	}
	return NewNodeWithDI("node-1", depends{dispatcher: d, store: store, getTimer: mockTimerFn, peersExplorer: newSimplePeersExplorer(peers), chatter: newMockChatter(), time: newMockTime(electionTimeSpan + 10)})
}
