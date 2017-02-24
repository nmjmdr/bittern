package raft


func beginElectionTimer(timerFn getTimerFn, dispatcher dispatcher, st *state) {
	timer := timerFn(getRandomElectionTimeout())
	go func() {
		_, ok := <-timer.Channel()
		if ok {
			dispatcher.dispatch(event{GotElectionSignal, st, nil})
		}
	}()
}
