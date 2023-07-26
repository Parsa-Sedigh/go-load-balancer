package strategy

import (
	"errors"
	"fmt"
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
)

/* Known load balancing strategies, each entry in this block should correspond to a load balancing strategy with a concrete implementation */
const (
	StrategyRoundRobin         = "RoundRobin"
	StrategyWeightedRoundRobin = "WeightedRoundRobin"
	StrategyUnknown            = "Unknown"
)

// BalancingStrategy is the load balancing abstraction that every load balancing algorithm should implement
type BalancingStrategy interface {
	// Every call to Next should give us the next server to forward the to
	Next([]*domain.Server) (*domain.Server, error)
}

type RoundRobin struct {
	//Servers *config.ServerList

	// the Current server to forward the request to.
	// the next server should be (current + 1) % len(Servers)
	Current uint32

	mu sync.Mutex
}

// Map of BalancingStrategy factories
var strategies map[string]func() BalancingStrategy

func init() {
	strategies = make(map[string]func() BalancingStrategy, 0)
	strategies[StrategyRoundRobin] = func() BalancingStrategy {
		return &RoundRobin{
			Current: 0,
			mu:      sync.Mutex{},
		}
	}

	strategies[StrategyWeightedRoundRobin] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
	}
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	//return (sl.current + 1) % uint32(len(sl.Servers))

	/* If more than one goroutine call Next at the same time, we could skip one of the live servers */
	r.mu.Lock()
	defer r.mu.Unlock()

	/* we keep incrementing seen until find a healthy server and then update current to be the value of `seen`, otherwise if we
	seen all of the servers, */
	seen := 0

	var picked *domain.Server

	for seen < len(servers) {
		picked = servers[r.Current]
		r.Current = (r.Current + 1) % uint32(len(servers))

		if picked.IsAlive() {
			break
		}

		seen++
	}

	if picked == nil || seen == len(servers) {
		log.Error("All servers are down")

		return nil, errors.New(fmt.Sprintf("Checked all the '%d' servers, none of them is available", seen))
	}

	//nxt := atomic.AddUint32(&r.Current, uint32(1))
	//lenS := uint32(len(servers))

	// wrap it around whatever number of services we have
	//if nxt >= lenS {
	//	nxt -= lenS
	//}

	//picked := servers[nxt%lenS]

	log.Infof("Strategy picked server '%s'", picked.Url.Host)

	return picked, nil
}

/*
	WeightedRoundRobin is a strategy that is similar to RoundRobin strategy, the only difference is that it takes server compute power

into consideration. The compute power of a server is given as an integer, it represents the fraction of requests that one server
can handle over another.

A RoundRobin is equivalent to WeightedRoundRobin strategy with all weights = 1
*/
type WeightedRoundRobin struct {
	// Any changes to the field below should only be done while holding the `mu` lock
	mu sync.Mutex

	/* This is making the assumption that the server list coming through the Next method, won't change between successive calls.
	Changing the server list would cause this strategy to break, panic or not route properly.

	count will keep track of the number of requests that the server `i` has processed.*/
	count []int

	/* current is the index of the last server that executed a request. */
	current int
}

func (w *WeightedRoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Is it first time using the strategy?
	if w.count == nil {
		w.count = make([]int, len(servers))
		w.current = 0
	}

	// represents the number of servers that are not alive till now
	seen := 0
	var picked *domain.Server

	for seen < len(servers) {
		picked = servers[w.current]
		capacity := picked.GetMetaOrDefaultInt("weight", 1)

		if !picked.IsAlive() {
			seen++

			/* Current server is not alive, so we reset the server's count and we try the next server in the next loop iteration(by
			incrementing the current field)*/
			w.count[w.current] = 0
			w.current = (w.current + 1) % len(servers)

			continue
		}

		if w.count[w.current] < capacity {
			w.count[w.current]++

			log.Infof("Strategy picked server '%s'", servers[w.current].Url.Host)

			return picked, nil
		}

		// server is at it's limit, reset the current server count and move on to the next server
		w.count[w.current] = 0
		w.current = (w.current + 1) % len(servers)
	}

	if picked == nil || seen == len(servers) {
		log.Error("All servers are down")

		return nil, errors.New(fmt.Sprintf("Checked all the '%d' servers, none of them is available", seen))
	}

	return picked, nil
}

/*
LoadStrategy will try and resolve the balancing strategy based on the name and will default to a 'StrategyRoundRobin' one if
no strategy matched.
*/
func LoadStrategy(name string) BalancingStrategy {
	st, ok := strategies[name]
	if !ok {
		log.Warnf("Strategy with name '%s' not found, falling back to the RoundRobin strategy", name)

		return strategies[StrategyRoundRobin]()
	}

	log.Infof("Picked strategy '%s'", name)

	return st()
}
