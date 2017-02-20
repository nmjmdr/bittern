package raft

type mockChatter struct {
  campaignStub func(peers []peer, currentTerm uint64)
}

func newMockChatter() *mockChatter {
  m := new(mockChatter);
  m.campaignStub = func (peers []peer, currentTerm uint64)  {
  }
  return m
}

func (m *mockChatter) campaign(peers []peer, currentTerm uint64) {
  m.campaignStub(peers,currentTerm)
}
