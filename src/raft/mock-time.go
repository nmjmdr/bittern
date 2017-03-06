package raft

type UnixNowCallbackFn func() int64
type mockTime struct {
	cb UnixNowCallbackFn
}

func newMockTime(unixNowFn UnixNowCallbackFn) *mockTime {
	m := new(mockTime)
	m.cb = unixNowFn
	return m
}

func (m *mockTime) UnixNow() int64 {
	return m.cb()
}
