package raft

type chatter interface {
	campaign(peers []peer, currentTerm uint64)
	sendVoteResponse(voteResponse voteResponse)
	sendAppendEntryResponse(entryResponse entryResponse)
}

/*
func parallelCampaigner(c *node) func(peers []peer, currentTerm uint64) {
	return func(peers []peer, currentTerm uint64) {
		vReq := voteRequest{c.Id(), currentTerm}
		for _, peer := range peers {
			go func() {
				vRes, err := c.d.transport.askVote(peer, vReq)
				if err != nil {
					// log and return
					return
				}
				if vRes.success {
					c.d.dispatcher.dispatch(event{GotVote, c.st, vRes})
				} else {
					c.d.dispatcher.dispatch(event{GotVoteRequestRejected, c.st, vRes})
				}
			}()
		}
	}
}
*/
