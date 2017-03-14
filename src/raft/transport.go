package raft

type Transport interface {
	SendVoteResponse(sendToPeer peer, vr voteResponse)
	SendAppendEntryResponse(sendToPeer peer, ar appendEntriesResponse)
	SendAppendEntriesRequest(peers []peer, ar appendEntriesRequest)
}
