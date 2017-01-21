package raft

// a dummy dispatcher used for testing
type directDispatcher struct {
	started bool
}

func (d *directDispatcher) start() {
	d.started = true
}

func (d *directDispatcher) dispatch(evt event) {
	switch evt.evtType {
	case GotElectionSignal:
		evt.st.stFn.gotElectionSignal(evt.st)
	}
}

func (d *directDispatcher) stop() {
	d.started = false
}
