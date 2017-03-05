package raft

type callbackFn func(event event)

type mockDispatcher struct {
	callback callbackFn
}

func newMockDispathcer(callback callbackFn) *mockDispatcher {
	m := new(mockDispatcher)
	m.callback = callback
	return m
}

func (m *mockDispatcher) Dispatch(event event) {

}
