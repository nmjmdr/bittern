package raft

type CampaignCallbackFn func(node *node)
type mockCampaigner struct {
	callback CampaignCallbackFn
}

func newMockCampaigner(callback CampaignCallbackFn) *mockCampaigner {
	m := new(mockCampaigner)
	m.callback = callback
	return m
}

func (m *mockCampaigner) Campaign(node *node) {
	if m.callback != nil {
		m.callback(node)
	}
}
