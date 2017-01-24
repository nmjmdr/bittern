package raft

type simplePeersExplorer struct {
	peers []peer
}

func newSimplePeersExplorer(peers []peer) *simplePeersExplorer {
	s := new(simplePeersExplorer)
	s.peers = peers
	return s
}

func (s *simplePeersExplorer) getPeers() []peer {
	return s.peers
}
