package domain

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
)

type Replica struct {
	Url      string            `yaml:"url"`
	Metadata map[string]string `yaml:"metadata"`
}

type Service struct {
	Name string `yaml:"name"`

	// A prefix matcher to select service based on the path part of the url
	/* Note(self): The matcher could be more sophisticated (i.e regex based, subdomain based), but for the purposes of simplicity, let's
	   think about this later.*/
	Matcher string `yaml:"matcher"`

	// Strategy is the load balancing strategy used for this service.
	Strategy string    `yaml:"strategy"`
	Replicas []Replica `yaml:"replicas"`
}

// Config is a representation of the configuration given to us from a config source
type Config struct {
	Services []Service `yaml:"services"`

	// Name of the strategy to be used in load balancing between instances
	Strategy string `yaml:"strategy"`
}

// Server is an instance of a running server
type Server struct {
	Url      *url.URL
	Proxy    *httputil.ReverseProxy
	Metadata map[string]string

	mu    sync.RWMutex
	alive bool
}

func (s *Server) Forward(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}

// GetMetaOrDefault returns the value associated with the given key in the metadata, or returns the default
func (s *Server) GetMetaOrDefault(key string, def string) string {
	v, ok := s.Metadata[key]
	if !ok {
		return def
	}

	return v
}

// GetMetaOrDefaultInt returns the int value associated with the given key in the metadata, or returns the default
func (s *Server) GetMetaOrDefaultInt(key string, def int) int {
	v := s.GetMetaOrDefault(key, fmt.Sprintf("%d", def))
	a, err := strconv.Atoi(v)
	if err != nil {
		return def
	}

	return a
}

// SetLiveness will change the current alive field value and return the old value.
func (s *Server) SetLiveness(value bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.alive
	s.alive = value

	return old
}

// IsAlive reports the liveness state of the server
func (s *Server) IsAlive() bool {
	s.mu.RLocker()
	defer s.mu.RUnlock()

	return s.alive
}
