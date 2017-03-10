package raft

type Transport interface {
	SendVoteResponse(sendToPeer peer, vr voteResponse)
}
