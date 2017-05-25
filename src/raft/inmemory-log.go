package raft

type inMemoryLog struct {
	buffer []entry
}

func newInmemoryLog() Log {
	i := new(inMemoryLog)
	// add a dummy first entry at index: 0
	i.buffer = make([]entry, 1)
	return i
}

func (i *inMemoryLog) LastTerm() uint64 {
	if len(i.buffer) > 1 {
		return i.buffer[len(i.buffer)-1].term
	}
	return 0
}

func (i *inMemoryLog) LastIndex() uint64 {
	if len(i.buffer) > 1 {
		return uint64(len(i.buffer) - 1)
	}
	return uint64(0)
}

func (i *inMemoryLog) EntryAt(index uint64) (entry, bool) {
	if len(i.buffer) > 1 && (index > uint64(0) && index < uint64(len(i.buffer))) {
		return i.buffer[index], true
	}
	return entry{}, false
}

func (i *inMemoryLog) AddAt(index uint64, entry entry) {
	if index <= 0 {
		return
	}
	if index > uint64(len(i.buffer)-1) {
		i.buffer = append(i.buffer, entry)
	} else {
		i.buffer[index] = entry
	}
}

func (i *inMemoryLog) DeleteFrom(index uint64) {
	if index <= 0 || index > uint64(len(i.buffer)-1) {
		return
	}
	i.buffer = i.buffer[0:index]
}

func (i *inMemoryLog) Append(entry entry) {
	i.buffer = append(i.buffer, entry)
}

func (i *inMemoryLog) Get(startIndex uint64) []entry {
	if startIndex <= 0 || startIndex > uint64(len(i.buffer)-1) || len(i.buffer) == 1 {
		panic("Index out of bounds of in-memory-log")
	}
	return i.buffer[startIndex:]
}
