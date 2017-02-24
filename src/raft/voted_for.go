package raft


func getVotedFor(term uint64,store store) (votedFor string) {
  votedFor, ok := store.getValue(votedForKey)
  if !ok {
    return ""
  }
  var inTerm uint64
  inTerm, ok = store.getInt(voteGrantedInTermKey)

  if !ok {
    panic("The term in which a vote was granted is not set")
  }

  if inTerm == term {
    return votedFor
  } else {
    return ""
  }

}

func setVotedFor(term uint64,votedFor string,store store) {
  store.setValue(votedForKey,votedFor)
  store.setInt(voteGrantedInTermKey,term)
}
