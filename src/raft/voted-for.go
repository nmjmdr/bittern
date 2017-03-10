package raft

type VotedFor interface {
	Get(term uint64) (candidateId string)
	Store(term uint64, candidateId string)
}

const LastVotedTermKey = "last-voted-term"
const VotedForKey = "voted-for"

type votedForStore struct {
	store Store
}

func newVotedForStore(store Store) *votedForStore {
	v := new(votedForStore)
	v.store = store
	return v
}

func (v *votedForStore) Get(term uint64) (candidateId string) {
	lastVotedTerm, ok := v.store.GetInt(LastVotedTermKey)
	if !ok || lastVotedTerm != term {
		return ""
	}
	candidateId = ""
	candidateId, ok = v.store.GetValue(VotedForKey)
	if !ok {
		panic("Could not obtain voted-for given that last voted term was set")
	}
	return candidateId
}

func (v *votedForStore) Store(term uint64, candidateId string) {
	v.store.StoreInt(LastVotedTermKey, term)
	v.store.StoreValue(VotedForKey, candidateId)
}
