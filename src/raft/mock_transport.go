package raft


type mockTransport struct {
}

func newMockTransport() transport {
	m := new(mockTransport)
	return m
}

func (m *mockTransport) askVote(peer peer, vr voteRequest) (voteResponse, error) {
	return voteResponse{},nil
}
