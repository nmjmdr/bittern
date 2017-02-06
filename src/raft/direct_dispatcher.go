package raft

import (
	"sync"
)
// a dummy dispatcher used for testing
type directDispatcher struct {
	started bool
	//dispatched chan bool
	signalled bool
	cond *sync.Cond
}

func newDirectDispatcher() *directDispatcher {
	d := new(directDispatcher)
	//d.dispatched = make(chan bool)
	d.cond = sync.NewCond(&sync.Mutex{})
	return d;
}

func (d *directDispatcher) start() {
	d.started = true
}

func (d *directDispatcher) dispatch(evt event) {
	switch evt.evtType {
	case GotElectionSignal:
		evt.st.stFn.gotElectionSignal()
	case GotVoteRequestRejected:
		evt.st.stFn.gotVoteRequestRejected(evt)
	case GotVote:
		evt.st.stFn.gotVote(evt)
	}
	// inform that the event was dispatched
	//d.dispatched <- true
	d.cond.L.Lock()
	d.signalled = true
	d.cond.Broadcast()
	d.cond.L.Unlock()
}

func (d *directDispatcher) awaitSignal() {
	d.cond.L.Lock()
	for !d.signalled {
		d.cond.Wait()
	}
	d.cond.L.Unlock()
}

func (d *directDispatcher) reset() {
	d.cond.L.Lock()
	d.signalled = false
	d.cond.L.Unlock()
}


func (d *directDispatcher) stop() {
	d.started = false
}
