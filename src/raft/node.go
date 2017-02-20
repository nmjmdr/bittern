package raft

import (
	"time"
)

type Node interface {
	Id() string
}

type depends struct {
	dispatcher    dispatcher
	store         store
	getTicker     getTickerFn
	peersExplorer peersExplorer
	campaigner    campaigner
}

type node struct {
	id string
	st *state
	d  depends
}

func (n *node) Id() string {
	return n.id
}

func NewNode(id string) Node {
	n := NewNodeWithDI(id, depends{
		dispatcher: newEventLoopdispatcher(),
		store:      nil,
		getTicker: func(d time.Duration) Ticker {
			return newBuiltInTicker(d)
		},
	},
	)
	return n
}

func NewNodeWithDI(id string, d depends) *node {
	n := new(node)
	n.id = id
	n.d = d
	n.st = n.getInitState()

	n.setCurrentTermOnBoot()

	n.d.dispatcher.start()
	// set the state function as a follower
	n.st.stFn = newFollower(n)
	return n
}

func (n *node) setCurrentTermOnBoot() {
	// set current term as 0, on boot for the first time
	_, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		n.d.store.setInt(currentTermKey, 0)
	}
}

func (n *node) getInitState() *state {
	st := new(state)
	st.commitIndex = 0 //?
	st.lastApplied = 0
	return st
}
