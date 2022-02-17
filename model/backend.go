package model

import "sync"

type Backend struct {
	URL    string `json:"url"`
	isDead bool
	mu     sync.Mutex
}

func (b *Backend) SetState(state bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isDead = state
}

func (b *Backend) IsDown() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	isAlive := b.isDead

	return isAlive
}
