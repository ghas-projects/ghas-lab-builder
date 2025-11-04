package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
)

// TokenManager manages token rotation across multiple goroutines using atomic operations
type TokenManager struct {
	tokens       []string
	currentIndex atomic.Int32
	logger       *slog.Logger
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(ctx context.Context, logger *slog.Logger) (*TokenManager, error) {
	tokens, ok := ctx.Value(config.TokenKey).([]string)
	if !ok || len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens found in context")
	}

	tm := &TokenManager{
		tokens: tokens,
		logger: logger,
	}
	// Always start at index 0
	tm.currentIndex.Store(0)

	return tm, nil
}

// GetCurrentToken returns the current token (thread-safe)
func (tm *TokenManager) GetCurrentToken() string {
	index := tm.currentIndex.Load()
	return tm.tokens[index]
}

// GetCurrentIndex returns the current token index (thread-safe)
func (tm *TokenManager) GetCurrentIndex() int {
	return int(tm.currentIndex.Load())
}

// RotateToken rotates to the next token atomically
// Returns error if no more tokens available
func (tm *TokenManager) RotateToken() error {
	// Atomically increment and get the new value
	newIndex := tm.currentIndex.Add(1)

	if int(newIndex) >= len(tm.tokens) {
		// We've exceeded the available tokens
		// Decrement back to stay at the last valid index
		tm.currentIndex.Add(-1)
		return fmt.Errorf("no more tokens available, already at index %d of %d", newIndex-1, len(tm.tokens)-1)
	}

	tm.logger.Info("Rotated to new token",
		slog.Int("new_index", int(newIndex)),
		slog.Int("total_tokens", len(tm.tokens)))

	return nil
}

// GetTokenCount returns the total number of tokens
func (tm *TokenManager) GetTokenCount() int {
	return len(tm.tokens)
}
