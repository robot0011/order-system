package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter holds the rate limiting data
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*RequestRecord
}

// RequestRecord holds information about a user's requests
type RequestRecord struct {
	Count    int
	LastTime time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*RequestRecord),
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

// CheckRateLimit checks if a user has exceeded the rate limit
func (rl *RateLimiter) CheckRateLimit(key string, maxRequests int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record, exists := rl.requests[key]

	if !exists || now.Sub(record.LastTime) > window {
		// New window, reset count
		rl.requests[key] = &RequestRecord{
			Count:    1,
			LastTime: now,
		}
		return true // Allowed
	}

	// Check if limit exceeded
	if record.Count >= maxRequests {
		return false // Rate limited
	}

	// Increment count
	record.Count++
	record.LastTime = now
	return true // Allowed
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, record := range rl.requests {
			if now.Sub(record.LastTime) > 10*time.Minute { // Keep records for 10 minutes
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// BruteForceProtector helps prevent brute force attacks
type BruteForceProtector struct {
	mu            sync.RWMutex
	attempts      map[string]*AttemptRecord
	lockoutPeriod time.Duration
}

// AttemptRecord holds information about login attempts
type AttemptRecord struct {
	Count     int
	LastTime  time.Time
	LockedOut bool
	LockTime  time.Time
}

// NewBruteForceProtector creates a new brute force protection instance
func NewBruteForceProtector() *BruteForceProtector {
	bfp := &BruteForceProtector{
		attempts:      make(map[string]*AttemptRecord),
		lockoutPeriod: 15 * time.Minute, // 15 minutes lockout
	}

	// Start cleanup goroutine
	go bfp.cleanup()

	return bfp
}

// CheckLoginAttempt checks if login attempt is allowed
func (bfp *BruteForceProtector) CheckLoginAttempt(username string) bool {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()

	now := time.Now()
	attempt, exists := bfp.attempts[username]

	if !exists {
		// First attempt
		bfp.attempts[username] = &AttemptRecord{
			Count:    1,
			LastTime: now,
		}
		return true
	}

	// Check if locked out
	if attempt.LockedOut {
		if now.Sub(attempt.LockTime) > bfp.lockoutPeriod {
			// Lockout period expired, reset
			bfp.resetAttempt(username)
			return true
		}
		return false // Still locked out
	}

	// Check if within time window (5 minutes)
	if now.Sub(attempt.LastTime) <= 5*time.Minute {
		if attempt.Count >= 5 { // Max 5 attempts in 5 minutes
			// Lock out the user
			attempt.LockedOut = true
			attempt.LockTime = now
			return false
		}
		// Increment attempt count
		attempt.Count++
	} else {
		// Reset after time window
		bfp.resetAttemptForUser(attempt, now)
	}

	return true
}

// RecordSuccessfulLogin removes failed attempts for the user
func (bfp *BruteForceProtector) RecordSuccessfulLogin(username string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()

	delete(bfp.attempts, username)
}

// resetAttemptForUser resets attempt record
func (bfp *BruteForceProtector) resetAttemptForUser(attempt *AttemptRecord, now time.Time) {
	attempt.Count = 1
	attempt.LastTime = now
	attempt.LockedOut = false
}

// resetAttempt resets attempt record
func (bfp *BruteForceProtector) resetAttempt(username string) {
	now := time.Now()
	bfp.attempts[username] = &AttemptRecord{
		Count:    1,
		LastTime: now,
	}
}

// cleanup removes old entries periodically
func (bfp *BruteForceProtector) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bfp.mu.Lock()
		now := time.Now()
		for username, attempt := range bfp.attempts {
			if attempt.LockedOut {
				if now.Sub(attempt.LockTime) > bfp.lockoutPeriod {
					delete(bfp.attempts, username)
				}
			} else if now.Sub(attempt.LastTime) > 10*time.Minute { // Clear inactive attempts after 10 minutes
				delete(bfp.attempts, username)
			}
		}
		bfp.mu.Unlock()
	}
}

// Global instances
var (
	rateLimiter         = NewRateLimiter()
	bruteForceProtector = NewBruteForceProtector()
)

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Use IP address as the key
		ip := c.IP()

		if !rateLimiter.CheckRateLimit(ip, maxRequests, window) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"error":   fmt.Sprintf("Rate limit exceeded. Please try again later."),
			})
		}

		return c.Next()
	}
}

// LoginRateLimitMiddleware applies specific rate limiting for login attempts
func LoginRateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Use IP address as the key
		ip := c.IP()

		// Allow max 10 login attempts per minute per IP
		if !rateLimiter.CheckRateLimit(ip, 10, time.Minute) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"error":   "Too many login attempts. Please try again later.",
			})
		}

		return c.Next()
	}
}

// BruteForceMiddleware adds brute force protection
func BruteForceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// For now, this is called in the login handler
		// We'll add specific checks there
		return c.Next()
	}
}

// CheckBruteForce attempts to check if a login attempt is allowed
func CheckBruteForce(username string) bool {
	return bruteForceProtector.CheckLoginAttempt(username)
}

// RecordSuccessfulLogin records a successful login
func RecordSuccessfulLogin(username string) {
	bruteForceProtector.RecordSuccessfulLogin(username)
}