package raft

const currentTermKey = "current-term"
const votedForKey = "voted-for"
const voteGrantedInTermKey = "vote-granted-in-term"
const nanos = 1000000
const electionTimeSpan = 150 * nanos
