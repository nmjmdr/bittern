package raft

type DispatchCallbackFn func(event event)

type mockDispatcher struct {
	callback DispatchCallbackFn
}

func newMockDispathcer() *mockDispatcher {
	m := new(mockDispatcher)
	return m
}

func (m *mockDispatcher) Dispatch(event event) {
	if m.callback != nil {
		m.callback(event)
	}
}
