package config

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Service struct {
	Name string `yaml:"name"`

	// A prefix matcher to select service based on the path part of the url
	/* Note(self): The matcher could be more sophisticated (i.e regex based, subdomain based), but for the purposes of simplicity, let's
	think about this later.*/
	Matcher  string   `yaml:"matcher"`
	Replicas []string `yaml:"replicas"`
}

// Config is a representation of the configuration given to us from a config source
type Config struct {
	Services []Service `yaml:"services"`

	// Name of the strategy to be used in load balancing between instances
	Strategy string `yaml:"strategy"`
}

// Server is an instance of a running server
type Server struct {
	Url   *url.URL
	Proxy *httputil.ReverseProxy
}

func (s *Server) Forward(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}

type ServerList struct {
	// Servers are the replicas
	Servers []*Server

	// Name of the service
	Name string
}
