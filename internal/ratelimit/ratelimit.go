// Package ratelimit provides algorithms for rate limiting requests
package ratelimit

// ClientInfo contains data that is needed to make a decision for rate limiting.
type ClientInfo struct {
	Identifier string // ip address, api key, etc
	Capacity   int
}

// Limiter defines an interface for rate limiting requests.
type Limiter interface {
	ClientAllowed(client ClientInfo) bool
}
