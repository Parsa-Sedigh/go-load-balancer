package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var (
	port = flag.Int("port", 8080, "where to start the load balancer")
)

type Service struct {
	Name     string
	Replicas []string
}

// Config is a representation of the configuration given to us from a config source
type Config struct {
	Services []Service

	// Name of the strategy to be used in load balancing between instances
	Strategy string
}

// Server is an instance of a running server
type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

type ServerList struct {
	Servers []*Server

	// the current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	current uint32
}

func (sl *ServerList) Next() uint32 {
	//return (sl.current + 1) % uint32(len(sl.Servers))
	nxt := atomic.AddUint32(&sl.current, uint32(1))
	lenS := uint32(len(sl.Servers))

	// wrap it around whatever number of services we have
	if nxt >= lenS {
		nxt -= lenS
	}

	return nxt
}

type LoadBalancer struct {
	Config     *Config
	ServerList *ServerList
}

func NewLoadBalancer(config *Config) *LoadBalancer {
	servers := make([]*Server, 0)

	for _, service := range config.Services {
		// TODO: Don't ignore the names
		for _, replica := range service.Replicas {
			url, err := url.Parse(replica)
			if err != nil {
				/* We don't want to continue creating a load balancer from malformed url for replicas. */
				log.Fatal(err)
			}

			proxy := httputil.NewSingleHostReverseProxy(url)
			servers = append(servers, &Server{
				url:   url,
				proxy: proxy,
			})
		}
	}

	return &LoadBalancer{
		Config: config,
		ServerList: &ServerList{
			Servers: servers,
			current: 0,
		},
	}
}

func (l *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: We need to support per service forwarding, i.e this method should read the request path, let's say
	// host:port/service/rest/of/url, this should be load balanced(forwarded) against service named "service" and the url will be
	// "host{i}:port{i}/rest/of/url"

	log.Info("Received new request: url='%s'", r.URL)

	nxt := l.ServerList.Next()

	// Load balancing(forwarding) the request to the proxy
	l.ServerList.Servers[nxt].proxy.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	conf := &Config{}
	loadBalancer := NewLoadBalancer(conf)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: loadBalancer,
	}

	if err := http.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
