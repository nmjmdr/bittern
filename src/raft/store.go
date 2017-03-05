package raft

type Store interface {
	getInt(key string) (uint64, bool)
	storeInt(key string, value uint64)
	getValue(key string) (string, bool)
	storeValue(key string, value string)
}
