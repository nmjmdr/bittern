package raft

type dispatchCallbackFn func(event event)

type mockDispatcher struct {
	callback dispatchCallbackFn
}

func newMockDispathcer(callback dispatchCallbackFn) *mockDispatcher {
	m := new(mockDispatcher)
	m.callback = callback
	return m
}

func (m *mockDispatcher) Dispatch(event event) {
  if m.callback != nil {
    m.callback(event)
  }
}
