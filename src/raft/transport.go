package raft

type Transport interface {
	SendVoteResponse(sendToPeer peer, vr voteResponse)
	SendAppendEntriesResponse(sendToPeer peer, ar appendEntriesResponse)
	SendAppendEntriesRequest(peers []peer, ar appendEntriesRequest)
}
