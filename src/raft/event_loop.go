package raft

type eventLoop struct {
	channel chan event
	done    chan bool
}

func newEventLoopdispatcher() dispatcher {
	l := new(eventLoop)
	l.done = make(chan bool)
	l.channel = make(chan event)

	return l
}

func (l *eventLoop) start() {
	go func(l *eventLoop) {
		for {
			select {
			case evt, ok := <-l.channel:
				if ok {
					l.mapToHandler(evt)
				}
			case <-l.done:
				return
			}
		}
	}(l)
}

func (l *eventLoop) mapToHandler(evt event) {
	switch evt.evtType {
	case GotElectionSignal:
		evt.st.stFn.gotElectionSignal()
	}
}

func (l *eventLoop) dispatch(evt event) {
	l.channel <- evt
}

func (l *eventLoop) stop() {
	l.done <- true
}
