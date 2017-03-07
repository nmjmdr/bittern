package raft

type Campaigner interface {
	Campaign(node *node)
}
