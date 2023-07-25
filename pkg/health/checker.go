package health

import (
	"errors"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type HealthChecker struct {
	servers []*domain.Server

	// TODO: configure the period based on the config file.
	period int
}

// NewChecker creates a new HealthChecker.
func NewChecker(conf *domain.Config, servers []*domain.Server) (*HealthChecker, error) {
	if len(servers) == 0 {
		return nil, errors.New("A server list expected, got an empty list")
	}

	return &HealthChecker{servers: servers}, nil
}

/*
Start keeps looping indefinitely trying to check the health of every server. The caller is responsible of creating
the goroutine when this should run.
*/
func (hc *HealthChecker) Start() {
	log.Info("Starting the health checker...")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, server := range hc.servers {
				go checkHealth(server)
			}
		}
	}
}

// checkHealth changes the liveness of the server(either from live to dead or the other way around)
func checkHealth(server *domain.Server) {
	/* We will consider a server to be healthy if we can open a TCP connection to the host:port of the server within a reasonable time frame. */
	_, err := net.DialTimeout("tcp", server.Url.Host, time.Second*5)
	if err != nil {
		log.Errorf("Could not connect to the server at '%s'", server.Url.Host)

		if old := server.SetLiveness(false); old {
			log.Warnf("Transitioning server '%s' from live to unavailable state", server.Url.Host)
		}

		return
	}

	if old := server.SetLiveness(true); !old {
		log.Infof("Transitioning server '%s' from unavailable to live state", server.Url.Host)
	}
}
