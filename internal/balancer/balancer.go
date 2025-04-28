// Package balancer contains algorithms for balancing incoming requests between backends.
package balancer

import (
	"fmt"
	"net/url"
)

// BackendServer represents a server, which accepts requests from load balancer.
type BackendServer struct {
	URL     *url.URL
	Healthy bool
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (*BackendServer, error)
	UpdateBackends(backends []*BackendServer)
}

// BackendServersFromArray creates an array of backend servers from config URLs.
func BackendServersFromArray(backends []string) ([]*BackendServer, error) {
	res := make([]*BackendServer, 0, len(backends))

	for _, b := range backends {
		parsedURL, err := url.Parse(b)
		if err != nil {
			return nil, fmt.Errorf("error parsing backend url: %w", err)
		}

		// all are healthy by default
		res = append(res, &BackendServer{URL: parsedURL, Healthy: true})
	}

	return res, nil
}
