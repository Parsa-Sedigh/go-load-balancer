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
)

var (
	port       = flag.Int("port", 8080, "where to start the load balancer")
	configFile = flag.String("config-path", "", "The config file to supply to load balancer")
)

type LoadBalancer struct {
	Config     *config.Config
	ServerList *config.ServerList
}

func NewLoadBalancer(conf *config.Config) *LoadBalancer {
	servers := make([]*config.Server, 0)

	for _, service := range conf.Services {
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
	}

	return &LoadBalancer{
		Config: conf,
		ServerList: &config.ServerList{
			Servers: servers,
			//Current: 0,
		},
	}
}

func (l *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: We need to support per service forwarding, i.e this method should read the request path, let's say
	// host:port/service/rest/of/url, this should be load balanced(forwarded) against service named "service" and the url will be
	// "host{i}:port{i}/rest/of/url"

	log.Infof("Received new request: url='%s'", r.Host)

	nxt := l.ServerList.Next()

	log.Infof("next: %d", nxt)

	// Load balancing(forwarding) the request to the proxy
	l.ServerList.Servers[nxt].Proxy.ServeHTTP(w, r)
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
