package raft

type WhoArePeers interface {
	All() []peer
}
