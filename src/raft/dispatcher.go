package raft

type dispatcher interface {
	start()
	dispatch(evt event)
	stop()
}
