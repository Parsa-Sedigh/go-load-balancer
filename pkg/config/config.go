package config

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Service struct {
	Name     string   `yaml:"name"`
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
	Servers []*Server

	// the Current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	Current uint32
}

func (sl *ServerList) Next() uint32 {
	//return (sl.current + 1) % uint32(len(sl.Servers))
	nxt := atomic.AddUint32(&sl.Current, uint32(1))
	lenS := uint32(len(sl.Servers))

	// wrap it around whatever number of services we have
	//if nxt >= lenS {
	//	nxt -= lenS
	//}

	return nxt % lenS
}
