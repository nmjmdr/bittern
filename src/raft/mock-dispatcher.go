package raft

type DispatchCallbackFn func(event event)

type mockDispatcher struct {
	callback DispatchCallbackFn
}

func newMockDispathcer(callback DispatchCallbackFn) *mockDispatcher {
	m := new(mockDispatcher)
	m.callback = callback
	return m
}

func (m *mockDispatcher) Dispatch(event event) {
	if m.callback != nil {
		m.callback(event)
	}
}
