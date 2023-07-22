package main

import (
	"flag"
	"fmt"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

var (
	port       = flag.Int("port", 8080, "where to start the load balancer")
	configFile = flag.String("config-path", "", "The config file to supply to load balancer")
)

type RoundRobin struct {
	Servers *config.ServerList

	// the Current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	Current uint32
}

// BalancingStrategy is the load balancing abstraction that every load balancing algorithm should implement
type BalancingStrategy interface {
	// Every call to Next should give us the next server to forward the to
	Next() (*config.Server, error)
}

func (r *RoundRobin) Next() (*config.Server, error) {
	//return (sl.current + 1) % uint32(len(sl.Servers))

	nxt := atomic.AddUint32(&r.Current, uint32(1))
	lenS := uint32(len(r.Servers.Servers))

	// wrap it around whatever number of services we have
	//if nxt >= lenS {
	//	nxt -= lenS
	//}

	return r.Servers.Servers[nxt%lenS], nil
}

type LoadBalancer struct {
	// Config is the configuration loaded from a config file
	/* TODO: This could be improved, as to fetch the configuration from a more abstract concept(like config source) that can either be a file
	or sth else and also should support hot reloading.*/
	Config *config.Config

	// ServerList will contain a mapping between matcher and replicas
	ServerList map[string]*config.ServerList
}

func NewLoadBalancer(conf *config.Config) *LoadBalancer {
	// TODO: Prevent multiple or invalid matchers before creating the server
	serverMap := make(map[string]*config.ServerList, 0)

	for _, service := range conf.Services {
		// for each service replica, we're gonna create a server instance

		servers := make([]*config.Server, 0)

		// TODO: Don't ignore the names
		for _, replica := range service.Replicas {
			url, err := url.Parse(replica)
			if err != nil {
				/* We don't want to continue creating a load balancer from malformed url for replicas. */
				log.Fatal(err)
			}

			proxy := httputil.NewSingleHostReverseProxy(url)

			servers = append(servers, &config.Server{
				Url:   url,
				Proxy: proxy,
			})
		}

		serverMap[service.Matcher] = &config.ServerList{
			Servers: servers,
			Name:    service.Name,
		}
	}

	return &LoadBalancer{
		Config:     conf,
		ServerList: serverMap,
	}
}

// findServiceList looks for the first server list that matches the reqPath(i.e matcher). Will return an error if no matcher
// have been found.
// TODO: Does it make sense to should allow default responders.
func (l *LoadBalancer) findServiceList(reqPath string) (*config.ServerList, error) {
	log.Infof("Trying to find matcher for request '%s'", reqPath)
	for matcher, s := range l.ServerList {
		if strings.HasPrefix(reqPath, matcher) {
			log.Infof("Found service '%s' matching the request", s.Name)

			return s, nil
		}
	}

	return nil, fmt.Errorf("Could not find a matcher for url: '%s", reqPath)
}

func (l *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: We need to support per service forwarding, i.e this method should read the request path, let's say
	// host:port/service/rest/of/url, this should be load balanced(forwarded) against service named "service" and the url will be
	// "host{i}:port{i}/rest/of/url"

	log.Infof("Received new request: url='%s'", r.Host)

	sl, err := l.findServiceList(r.URL.Path)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)

		return
	}

	nxt := sl.Next()

	log.Infof("next: %d", nxt)

	// Load balancing(forwarding) the request to the proxy
	sl.Servers[nxt].Proxy.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	//conf := &config.Config{
	//	Services: []config.Service{
	//		{
	//			Name:     "Test",
	//			Replicas: []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
	//		},
	//	},
	//}

	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	conf, err := config.LoadConfig(file)
	if err != nil {
		log.Fatal(err)
	}

	loadBalancer := NewLoadBalancer(conf)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: loadBalancer,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
