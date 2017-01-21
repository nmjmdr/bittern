package raft

func beginElectionTimer(tickerFn getTickerFn, dispatcher dispatcher, st *state) {
	ticker := tickerFn(getRandomElectionTimeout())
	go func() {
		_, ok := <-ticker.Channel()
		if ok {
			dispatcher.dispatch(event{GotElectionSignal, st, nil})
		}
	}()
}
