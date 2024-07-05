// app/middleware.go
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter tracks the visitors and their request counts
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

// visitor represents a visitor with the last seen time and request count
type visitor struct {
	lastSeen   time.Time
	requests   int
	timestamps []time.Time
	mu         sync.Mutex
}

// Initialize a new RateLimiter
func newRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanupVisitors()
	return rl
}

// Get or create a visitor entry for an IP address
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{lastSeen: time.Now()}
		rl.visitors[ip] = v
	}
	return v
}

// Periodically clean up old visitors
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware to apply rate limiting based on the form configuration
func rateLimitMiddleware(next http.Handler, rl *RateLimiter, config Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formID := r.FormValue("formid")
		if formID == "" {
			http.Error(w, "Form ID is required", http.StatusBadRequest)
			log.Warn("Form ID is required")
			return
		}

		formConfig, exists := config.Forms[formID]
		if !exists {
			http.Error(w, "Form configuration not found", http.StatusBadRequest)
			log.Warnf("Form configuration not found for ID: %s", formID)
			return
		}

		duration, err := time.ParseDuration(formConfig.RateLimit.Duration)
		if err != nil {
			http.Error(w, "Invalid rate limit duration", http.StatusInternalServerError)
			log.Errorf("Invalid rate limit duration for form %s: %v", formID, err)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid IP address", http.StatusInternalServerError)
			log.Errorf("Invalid IP address: %v", err)
			return
		}

		visitor := rl.getVisitor(ip)
		visitor.mu.Lock()
		defer visitor.mu.Unlock()

		log.Infof("Visitor %s - requests: %d, lastSeen: %s", ip, visitor.requests, visitor.lastSeen)
		if time.Since(visitor.lastSeen) > duration {
			visitor.requests = 0
			visitor.timestamps = []time.Time{}
		}

		if visitor.requests >= formConfig.RateLimit.Requests {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			log.Warnf("Rate limit exceeded for IP: %s, form ID: %s", ip, formID)
			return
		}

		visitor.requests++
		visitor.lastSeen = time.Now()
		visitor.timestamps = append(visitor.timestamps, visitor.lastSeen)
		log.Infof("Visitor %s - incremented requests to: %d", ip, visitor.requests)

		next.ServeHTTP(w, r)
	})
}

// Middleware to handle authentication
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		auth, ok := session.Values["authenticated"].(bool)
		log.Infof("Auth check - authenticated: %v, ok: %v", auth, ok)
		if !ok || !auth {
			log.Warn("Unauthorized access attempt")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Generate a random string of given length
func generateRandomString(n int) (string, error) {
	const letters = "0123456789"
	bytes := make([]byte, n)
	for i := range bytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Errorf("Error generating random string: %v", err)
			return "", err
		}
		bytes[i] = letters[num.Int64()]
	}
	return string(bytes), nil
}

// Validate the uploaded file based on field configuration
func validateFile(handler *multipart.FileHeader, field Field) error {
	if handler.Size > field.MaxFileSize {
		return fmt.Errorf("file size exceeds the maximum allowed size of %d bytes", field.MaxFileSize)
	}

	fileType := handler.Header.Get("Content-Type")
	validType := false
	for _, allowedType := range field.AllowedFileTypes {
		if fileType == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("file type %s is not allowed", fileType)
	}

	return nil
}

// Middleware to handle dynamic CORS based on form configuration
func dynamicCORSMiddleware(next http.Handler, config Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formID := r.FormValue("formid")
		if formID == "" {
			http.Error(w, "Form ID is required", http.StatusBadRequest)
			log.Warn("Form ID is required")
			return
		}

		formConfig, exists := config.Forms[formID]
		if !exists {
			http.Error(w, "Form configuration not found", http.StatusBadRequest)
			log.Warnf("Form configuration not found for ID: %s", formID)
			return
		}

		// Check the referral URL
		referer := r.Referer()
		if referer == "" || !strings.HasPrefix(referer, formConfig.ReferralURL) {
			http.Error(w, "Invalid referral URL", http.StatusForbidden)
			log.Warnf("Invalid referral URL: %s", referer)
			return
		}

		// Check CORS origins
		origin := r.Header.Get("Origin")
		if origin == "" {
			http.Error(w, "Origin header is required", http.StatusForbidden)
			log.Warn("Origin header is required")
			return
		}

		allowed := false
		for _, allowedOrigin := range formConfig.AllowedOrigins {
			if strings.EqualFold(allowedOrigin, origin) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "CORS not allowed for this origin", http.StatusForbidden)
			log.Warnf("CORS not allowed for origin: %s", origin)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
