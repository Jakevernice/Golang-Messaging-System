package auth

import (
	"sync"
	"time"
)

// TokenBlacklist manages a list of invalidated tokens
type TokenBlacklist struct {
	blacklist     map[string]time.Time // Maps token string to expiration time
	mutex         sync.RWMutex
	cleanupTicker *time.Ticker
}

// NewTokenBlacklist creates a new token blacklist with automatic cleanup
func NewTokenBlacklist(cleanupInterval time.Duration) *TokenBlacklist {
	tb := &TokenBlacklist{
		blacklist:     make(map[string]time.Time),
		cleanupTicker: time.NewTicker(cleanupInterval),
	}

	// Start cleanup goroutine
	go tb.periodicCleanup()

	return tb
}

// Add adds a token to the blacklist with a specified TTL (time to live)
func (tb *TokenBlacklist) Add(token string, expiry time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.blacklist[token] = expiry
}

// IsBlacklisted checks if a token is in the blacklist
func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	expiryTime, exists := tb.blacklist[token]
	if !exists {
		return false
	}

	// If token has expired, we can remove it from the blacklist
	if time.Now().After(expiryTime) {
		// Use defer to avoid deadlock when upgrading from read lock to write lock
		defer func() {
			tb.mutex.Lock()
			delete(tb.blacklist, token)
			tb.mutex.Unlock()
		}()
		return false
	}

	return true
}

// cleanup removes expired tokens from the blacklist
func (tb *TokenBlacklist) cleanup() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	for token, expiry := range tb.blacklist {
		if now.After(expiry) {
			delete(tb.blacklist, token)
		}
	}
}

// periodicCleanup runs the cleanup function at regular intervals
func (tb *TokenBlacklist) periodicCleanup() {
	for range tb.cleanupTicker.C {
		tb.cleanup()
	}
}

// Stop stops the cleanup ticker
func (tb *TokenBlacklist) Stop() {
	tb.cleanupTicker.Stop()
}

// Global token blacklist instance
var (
	globalBlacklist     *TokenBlacklist
	globalBlacklistOnce sync.Once
)

// GetTokenBlacklist returns the global token blacklist instance
func GetTokenBlacklist() *TokenBlacklist {
	globalBlacklistOnce.Do(func() {
		// Clean up every hour by default
		globalBlacklist = NewTokenBlacklist(1 * time.Hour)
	})
	return globalBlacklist
}
