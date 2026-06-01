package cache

import (
	"sync"
	"time"
)

type PendingRegistration struct {
	Email          string
	FirstName      string
	LastName       string
	Username       string
	HashedPassword string
	OTPHash        string
	ExpiresAt      time.Time
	Attempts       int
}

type RegistrationCache struct {
	mu    sync.RWMutex
	store map[string]PendingRegistration
}

func NewRegistrationCache() *RegistrationCache {
	return &RegistrationCache{
		store: make(map[string]PendingRegistration),
	}
}

func (c *RegistrationCache) Set(email string, data PendingRegistration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[email] = data
}

func (c *RegistrationCache) Get(email string) (PendingRegistration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.store[email]
	return data, ok
}

func (c *RegistrationCache) Update(email string, data PendingRegistration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[email] = data
}

func (c *RegistrationCache) Delete(email string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, email)
}

// StartCleanup runs a background goroutine that removes expired entries every 5 minutes.
func (c *RegistrationCache) StartCleanup(done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				c.mu.Lock()
				for email, data := range c.store {
					if data.ExpiresAt.Before(now) {
						delete(c.store, email)
					}
				}
				c.mu.Unlock()
			case <-done:
				return
			}
		}
	}()
}
