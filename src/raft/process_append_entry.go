package raft

import (
	"fmt"
)

func processAppendEntry(node *node, entryRequest *entryRequest) (entryAccepted bool) {
	currentTerm, ok := node.d.store.getInt(currentTermKey)
	if !ok {
		panic("Unable to obtain current term from store")
	}

	// reject if the term is less
	if entryRequest.term < currentTerm || !checkLog() {
		// reject, send a reply saying it was rejected
		node.d.chatter.sendAppendEntryResponse(entryResponse{success: false, term: currentTerm})
		return false
	}
	// set last heard from leader
	node.lastHeardFromALeader = node.d.time.unixNano()
	// replicate the log here, later
	fmt.Println("TO DO: replicate the log here!")
	node.d.chatter.sendAppendEntryResponse(entryResponse{success: true, term: currentTerm})
	return true
}
