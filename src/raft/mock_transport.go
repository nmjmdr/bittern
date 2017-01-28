package raft

type askVoteCallbackFn func(peer peer, vr voteRequest) (voteResponse, error)

type mockTransport struct {
	askVoteCb askVoteCallbackFn
}

func newMockTransport(askVoteCb askVoteCallbackFn) transport {
	m := new(mockTransport)
	m.askVoteCb = askVoteCb
	return m
}

func (m *mockTransport) askVote(peer peer, vr voteRequest) (voteResponse, error) {
	return m.askVoteCb(peer, vr)
}
