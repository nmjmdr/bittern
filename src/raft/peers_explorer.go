package raft

type peer struct {
	Id      string
	Address string
}

type peersExplorer interface {
	getPeers() []peer
}
