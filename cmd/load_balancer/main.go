package main

import (
	"flag"
	"fmt"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/config"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/domain"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/health"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/strategy"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var (
	port       = flag.Int("port", 8080, "where to start the load balancer")
	configFile = flag.String("config-path", "", "The config file to supply to load balancer")
)

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

		servers := make([]*domain.Server, 0)

		// TODO: Don't ignore the names
		for _, replica := range service.Replicas {
			url, err := url.Parse(replica.Url)
			if err != nil {
				/* We don't want to continue creating a load balancer from malformed url for replicas. */
				log.Fatal(err)
			}

			proxy := httputil.NewSingleHostReverseProxy(url)

			servers = append(servers, &domain.Server{
				Url:      url,
				Proxy:    proxy,
				Metadata: replica.Metadata,
			})
		}

		checker, err := health.NewChecker(nil, servers)
		if err != nil {
			log.Fatal(err)
		}

		serverMap[service.Matcher] = &config.ServerList{
			Servers:  servers,
			Name:     service.Name,
			Strategy: strategy.LoadStrategy(service.Strategy),
			HC:       checker,
		}
	}

	// start all the health checkers for all the provided matchers
	for _, sl := range serverMap {
		go sl.HC.Start()
	}

	return &LoadBalancer{
		Config:     conf,
		ServerList: serverMap,
	}
}

// findServiceList looks for the first server list that matches the reqPath(matcher). Will return an error if no matcher
// have been found.
// TODO: Does it make sense to allow default responders?
func (l *LoadBalancer) findServiceList(reqPath string) (*config.ServerList, error) {
	log.Infof("Trying to find matcher for request '%s'", reqPath)

	for matcher, s := range l.ServerList {
		if strings.HasPrefix(reqPath, matcher) {
			log.Infof("Found service '%s' matching the request", s.Name)

			return s, nil
		}
	}

	return nil, fmt.Errorf("could not find a matcher for url: '%s", reqPath)
}

func (l *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: We need to support per service forwarding, in other words, this method should read the request path, let's say
	// host:port/service/rest/of/url, this should be load balanced(forwarded) against service named "service" and the url will be
	// "host{i}:port{i}/rest/of/url"

	log.Infof("Received new request: url='%s'", r.Host)

	sl, err := l.findServiceList(r.URL.Path)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)

		return
	}

	next, err := sl.Strategy.Next(sl.Servers)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	log.Infof("Forwarding to the server = '%s'", next.Url.Host)

	// Load balancing and forwarding the request to the proxy
	//sl.Servers[next].Proxy.ServeHTTP(w, r)
	next.Forward(w, r)
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
