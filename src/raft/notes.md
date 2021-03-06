Notes:

1>
log is at least as up-to-date as receiver’s log, grant vote
Raft determines which of two logs is more up-to-date by comparing the index and term of
the last entries in the logs. If the logs have last entries with different terms,
then the log with the later term is more up-to-date. If the logs end with the same term,
then whichever log is longer is more up-to-date.

2>
what about a leader stepping down in append entries call?, the argument here could be that, if there is a leader
who was isolated and rejoined the network, it would step-down, when it sends its append entry and
realizes that it is no longer the up to date leader
Not required to check for a leader if it has step down in append entries

3>
// have we already voted for another peer in request's term?
// If votedFor is empty or candidateId, then grant vote
// where votedFor => candidateId that received vote in current term (or null if none)
// who did we vote for in the term?

4> This has to be done for apply to state machine:
If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)

5>
If successful: update nextIndex and matchIndex for follower (§5.3)
If AppendEntries fails because of log inconsistency: decrement nextIndex and retry (§5.3)
If there exists an N such that N > commitIndex, a majority of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N (§5.3, §5.4).

6> matchIndex vs nextIndex
One common source of confusion is the difference between nextIndex and matchIndex. In particular, you may observe that
matchIndex = nextIndex - 1, and simply not implement matchIndex. This is not safe.

While nextIndex and matchIndex are generally updated at the same time to a similar value (specifically, nextIndex = matchIndex + 1),
the two serve quite different purposes. nextIndex is a guess as to what prefix the leader shares with a given follower.
It is generally quite optimistic (we share everything), and is moved backwards only on negative responses.
For example, when a leader has just been elected, nextIndex is set to be index at the end of the log.
In a way, nextIndex is used for performance – you only need to send these things to this peer.

matchIndex is used for safety. It is a conservative measurement of what prefix of the log the
leader shares with a given follower. matchIndex cannot ever be set to a value that is too high,
as this may cause the commitIndex to be moved too far forward.
This is why matchIndex is initialized to -1 (i.e., we agree on no prefix),
and only updated when a follower positively acknowledges an AppendEntries RPC.

7> DO have to incorporate the concept of outstanding requests: why?
Let us assume we are yet to receive the response to first outstanding append entry request
If we send another one: it would include the entries from previous request as well?
For now: let us not complicate and follow a simple approach: we always send

8> Look at this for understanding of commiting in the presence of network partition: https://thesecretlivesofdata.com/raft/
Especially: if a leader cannot commit to majority of the nodes, it should stay uncommitted
Reference: point 5> in notes, the following rule ensures it:
  If there exists an N such that N > commitIndex, a majority of matchIndex[i] >= N and log[N].term == currentTerm:
  then set commitIndex = N
  Only start from the current commitIndex and then start checking the next one
  If a leader is not able to send heartbeat or append entry to majority of the nodes within an election timeout
  then it has to step down

Links:
https://groups.google.com/d/topic/raft-dev/M4BraEWG3TU/discussion
https://thesquareplanet.com/blog/students-guide-to-raft/
