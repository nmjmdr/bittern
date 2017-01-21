package raft

type store interface {
	getInt(key string) (uint64, bool)
	getValue(key string) (string, bool)
	setInt(key string, val uint64)
	setValue(key string, val string)
}
