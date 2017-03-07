package raft

type AllPeersCallbackFn func() []peer
type mockWhoArePeers struct {
  callback AllPeersCallbackFn
}

func newMockWhoArePeers(callback AllPeersCallbackFn) *mockWhoArePeers {
  m := new(mockWhoArePeers)
  m.callback = callback
  return m
}

func (m *mockWhoArePeers) All() []peer {
  return m.callback()
}
