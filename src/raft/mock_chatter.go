package raft

type mockChatter struct {
  campaignStub func(peers []peer, currentTerm uint64)
  sendVoteResponseStub func (voteResponse voteResponse)
}

func newMockChatter() *mockChatter {
  m := new(mockChatter);
  m.campaignStub = func (peers []peer, currentTerm uint64)  {
  }
  m.sendVoteResponseStub = func(voteResponse voteResponse) {
  }
  return m
}

func (m *mockChatter) campaign(peers []peer, currentTerm uint64) {
  m.campaignStub(peers,currentTerm)
}

func (m *mockChatter) sendVoteResponse(voteResponse voteResponse) {
  m.sendVoteResponseStub(voteResponse)
}
