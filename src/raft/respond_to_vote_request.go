package raft

func respondToVoteRequest(evt event, n *node) {
	request := evt.payload.(*voteRequest)
	currentTerm, ok := n.d.store.getInt(currentTermKey)
	if !ok {
		panic("Could not obtain current term key in candidate")
	}
	if request.term < currentTerm {
		n.d.chatter.sendVoteResponse(voteResponse{Success: false, Term: currentTerm, From: n.id})
		return
	}

	votedFor := getVotedFor(request.term, n.d.store)

	if (votedFor == "" || votedFor == request.from) && checkCandidatesLog() {
		n.d.chatter.sendVoteResponse(voteResponse{Success: true, Term: currentTerm, From: n.id})
		setVotedFor(request.term, request.from, n.d.store)
		return
	}

}
