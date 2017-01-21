package raft

type inMemorystore struct {
	m map[string](interface{})
}

func newInMemorystore() store {
	i := new(inMemorystore)
	i.m = make(map[string](interface{}))
	return i
}

func (i *inMemorystore) getInt(key string) (uint64, bool) {
	v, ok := i.m[key]
	if !ok {
		return 0, false
	}
	return v.(uint64), true
}

func (i *inMemorystore) getValue(key string) (string, bool) {
	v, ok := i.m[key]
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (i *inMemorystore) setInt(key string, val uint64) {
	i.m[key] = val
}

func (i *inMemorystore) setValue(key string, val string) {
	i.m[key] = val
}
