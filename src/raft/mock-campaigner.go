package raft

type CampaignCallbackFn func(node *node)
type mockCampaigner struct {
	callback CampaignCallbackFn
}

func newMockCampaigner() *mockCampaigner {
	m := new(mockCampaigner)
	return m
}

func (m *mockCampaigner) Campaign(node *node) {
	if m.callback != nil {
		m.callback(node)
	}
}
