package raft

type SendVoteResponseCallbackFn func(sendToPeer peer, voteResponse voteResponse)
type SendAppendEntryResponseCallbackFn func(sendToPeer peer,ar appendEntryResponse)
type mockTransport struct {
	sendVoteResponseCb SendVoteResponseCallbackFn
  sendAppendEntryResponseCb SendAppendEntryResponseCallbackFn
}

func newMockTransport() *mockTransport {
	t := new(mockTransport)
	return t
}

func (t *mockTransport) SendVoteResponse(sendToPeer peer, vr voteResponse) {
	if t.sendVoteResponseCb != nil {
		t.sendVoteResponseCb(sendToPeer, vr)
	}
}

func (t *mockTransport) SendAppendEntryResponse(sendToPeer peer,ar appendEntryResponse) {
  if t.sendAppendEntryResponseCb != nil {
    t.sendAppendEntryResponseCb(sendToPeer,ar)
  }
}
