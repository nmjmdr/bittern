package raft

type SendVoteResponseCallbackFn func(sendToPeer peer, voteResponse voteResponse)
type mockTransport struct {
	sendVoteResponseCb SendVoteResponseCallbackFn
}

func newMockTransport(sendVoteResponseCb SendVoteResponseCallbackFn) *mockTransport {
	t := new(mockTransport)
	t.sendVoteResponseCb = sendVoteResponseCb
	return t
}

func (t *mockTransport) SendVoteResponse(sendToPeer peer, vr voteResponse) {
	if t.sendVoteResponseCb != nil {
		t.sendVoteResponseCb(sendToPeer, vr)
	}
}
