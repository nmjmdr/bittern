package raft

type Store interface {
	GetInt(key string) (uint64, bool)
	StoreInt(key string, value uint64)
	getValue(key string) (string, bool)
	StoreValue(key string, value string)
}
