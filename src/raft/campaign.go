package raft

import (
	"fmt"
)

type askVotesFn func(peers []peer)

func getCampaign(transport transport, dispatcher dispatcher, st *state,vr voteRequest) askVotesFn {
	// ask votes in parallel from each peer
	return func(peers []peer) {
		for _, peer := range peers {
			go func() {
				voteResponse, err := transport.askVote(peer,vr)
				if err != nil {
					// could not communicate with peer, log and ignore
					fmt.Printf("Could not communicate with peer: %s", peer.Id)
					return
				}
				dispatchvoteResponse(transport, dispatcher, st, voteResponse)
			}()
		}
	}
}

func dispatchvoteResponse(transport transport, dispatcher dispatcher, st *state, voteResponse voteResponse) {
	if voteResponse.Success {
		dispatcher.dispatch(event{GotVote, st, voteResponse})
	} else {
		dispatcher.dispatch(event{GotVote, st, voteResponse})
	}
}
