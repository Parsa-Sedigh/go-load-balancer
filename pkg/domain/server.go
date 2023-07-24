package domain

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Replica struct {
	Url      string            `yaml:"url"`
	metadata map[string]string `yaml:"metadata"`
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
type Server[m string] struct {
	Url      *url.URL
	Proxy    *httputil.ReverseProxy
	Metadata map[string]m
}

func (s *Server[]) Forward(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}

// GetMetaOrDefault returns the value associated with the given key in the metadata, or returns the default
func (s *Server[E]) GetMetaOrDefault(key, def E) E {
	v, ok := s.Metadata[key]
	if !ok {
		return def
	}

	return v
}
