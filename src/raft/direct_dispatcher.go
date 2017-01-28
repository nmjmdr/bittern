package raft

// a dummy dispatcher used for testing
type directDispatcher struct {
	started bool
	dispatched chan bool
}

func newDirectDispatcher() *directDispatcher {
	d := new(directDispatcher)
	d.dispatched = make(chan bool)
	return d;
}

func (d *directDispatcher) start() {
	d.started = true
}

func (d *directDispatcher) dispatch(evt event) {
	switch evt.evtType {
	case GotElectionSignal:
		evt.st.stFn.gotElectionSignal()
	}
	// inform that the event was dispatched
	d.dispatched <- true
}

func (d *directDispatcher) stop() {
	d.started = false
}
