package raft

type SendVoteResponseCallbackFn func(sendToPeer peer, voteResponse voteResponse)
type SendAppendEntriesResponseCallbackFn func(sendToPeer peer, ar appendEntriesResponse)
type SendAppendEntriesRequestCallbackFn func(peer []peer, ar appendEntriesRequest)
type mockTransport struct {
	sendVoteResponseCb          SendVoteResponseCallbackFn
	sendAppendEntriesResponseCb SendAppendEntriesResponseCallbackFn
	sendAppendEntriesRequestCb  SendAppendEntriesRequestCallbackFn
}

func newMockTransport() *mockTransport {
	t := new(mockTransport)
	return t
}

func (t *mockTransport) SendVoteResponse(sendToPeer peer, vr voteResponse) {
	if t.sendVoteResponseCb != nil {
		t.sendVoteResponseCb(sendToPeer, vr)
		return
	}
	panic("sendVoteResponseCb was not set, but mockTransport.SendVoteResponse was invoked")
}

func (t *mockTransport) SendAppendEntryResponse(sendToPeer peer, ar appendEntriesResponse) {
	if t.sendAppendEntriesResponseCb != nil {
		t.sendAppendEntriesResponseCb(sendToPeer, ar)
		return
	}
	panic("sendAppendEntriesResponseCb was not set, but mockTransport.SendAppendEntryResponse was invoked")
}

func (t *mockTransport) SendAppendEntriesRequest(peers []peer, ar appendEntriesRequest) {
	if t.sendAppendEntriesRequestCb != nil {
		t.sendAppendEntriesRequestCb(peers, ar)
		return
	}
	panic("sendAppendEntriesRequestCb was not set, but mockTransport.SendAppendEntriesRequest was invoked")
}
