package raft

type inMemoryStore struct {
	m map[string]interface{}
}

func newInMemoryStore() *inMemoryStore {
	inMem := new(inMemoryStore)
	inMem.m = make(map[string]interface{})
	return inMem
}

func (i *inMemoryStore) GetInt(key string) (uint64, bool) {
	value, ok := i.m[key]
	if !ok {
		return uint64(0), false
	}
	return value.(uint64), true
}

func (i *inMemoryStore) StoreInt(key string, value uint64) {
	i.m[key] = value
}
func (i *inMemoryStore) getValue(key string) (string, bool) {
	value, ok := i.m[key]
	if !ok {
		return "", false
	}
	return value.(string), true
}
func (i *inMemoryStore) StoreValue(key string, value string) {
	i.m[key] = value
}
