package raft

import (
  "testing"
)


func createNode() *node {
  n := newNode()
  n.dispatcher = newMockDispathcer()
  return n
}

func Test_WhenTheNodeBootsItShouldStartAsAFollower(t *testing.T) {
  n := createNode()
  n.boot()
  if n.st.mode != Follower {
    t.Fatal("Should have initialized as a follower")
  }
}

func Test_WhenTheNodeBootsItShouldSetTheTermToZero(t *testing.T) {
  n := createNode()
  n.boot()
  if n.st.term != 0 {
    t.Fatal("Should have initialized the term to 0")
  }
}
