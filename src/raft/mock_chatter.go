package raft

type mockChatter struct {
	campaignStub                func(peers []peer, currentTerm uint64)
	sendVoteResponseStub        func(voteResponse voteResponse)
	sendAppendEntryResponseStub func(entryResponse entryResponse)
}

func newMockChatter() *mockChatter {
	m := new(mockChatter)
	m.campaignStub = func(peers []peer, currentTerm uint64) {
	}
	m.sendVoteResponseStub = func(voteResponse voteResponse) {
	}
	m.sendAppendEntryResponseStub = func(entryResponse entryResponse) {
	}
	return m
}

func (m *mockChatter) campaign(peers []peer, currentTerm uint64) {
	m.campaignStub(peers, currentTerm)
}

func (m *mockChatter) sendVoteResponse(voteResponse voteResponse) {
	m.sendVoteResponseStub(voteResponse)
}

func (m *mockChatter) sendAppendEntryResponse(entryResponse entryResponse) {
	m.sendAppendEntryResponseStub(entryResponse)
}
