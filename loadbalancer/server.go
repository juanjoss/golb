package loadbalancer

import "sync"

type server struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	isDead bool
	mu     sync.Mutex
}

func (b *server) SetState(state bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isDead = state
}

func (b *server) IsDown() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	isAlive := b.isDead

	return isAlive
}
