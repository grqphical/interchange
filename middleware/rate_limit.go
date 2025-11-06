package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/grqphical/interchange/templates"
	"github.com/spf13/viper"
)

type RateLimitDatabase struct {
	Buckets map[string]int
	mu      sync.Mutex
}

func (r *RateLimitDatabase) Update() {
	r.mu.Lock()
	for ip := range r.Buckets {
		r.Buckets[ip] += 1
	}
	r.mu.Unlock()
}

func (r *RateLimitDatabase) CheckRateLimit(remoteAddr string) bool {
	r.mu.Lock()
	remoteAddrComponents := strings.Split(remoteAddr, ":")
	ip := remoteAddrComponents[0]

	if tokens, exists := r.Buckets[ip]; exists {
		if tokens == 0 {
			r.mu.Unlock()
			return false
		} else {
			r.Buckets[ip] = tokens - 1
			r.mu.Unlock()
			return true
		}
	} else {
		r.Buckets[ip] = viper.GetInt("rate_limiting.max_requests") - 1
		r.mu.Unlock()
		return true
	}
}

var RateLimitDB RateLimitDatabase

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RateLimitDB.CheckRateLimit(r.RemoteAddr) {
			next.ServeHTTP(w, r)
		} else {
			templates.WriteError(w, http.StatusTooManyRequests, "You have been rate limited. Please try again later")
		}
	})
}

func init() {
	RateLimitDB = RateLimitDatabase{
		Buckets: make(map[string]int),
	}

	go RateLimitDB.Update()
}
