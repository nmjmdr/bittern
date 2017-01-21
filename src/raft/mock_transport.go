package raft

type askForVoteCallbackFn func(peer peer,vr voteRequest) (voteResponse,error)

type callbackMap struct {
	askForVote askForVoteCallbackFn
}

type mockTransport struct {
	cbMap *callbackMap
}

func newMockTransport(cbMap *callbackMap) transport {
	m := new(mockTransport)
	m.cbMap = cbMap
	return m
}

func (m *mockTransport) askVote(peer peer,vr voteRequest) (voteResponse, error) {
	return m.cbMap.askForVote(peer,vr)
}
