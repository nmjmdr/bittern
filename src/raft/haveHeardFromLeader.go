package raft



func haveHeardFromALeader(time Time,lastHeardFromLeader int64) bool {
  now := time.unixNano()
  return ( (now - lastHeardFromLeader) < electionTimeSpan)
}
