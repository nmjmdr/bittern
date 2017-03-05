package raft

type mockDispatcher struct {
}

func newMockDispathcer() *mockDispatcher {
  m := new(mockDispatcher)
  return m
}

func (m *mockDispatcher) Dispatch(event event) {
  
}
