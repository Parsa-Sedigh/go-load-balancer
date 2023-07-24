package strategy

import (
	"github.com/Parsa-Sedigh/go-load-balancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
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
}

// Map of BalancingStrategy factories
var strategies map[string]func() BalancingStrategy

func init() {
	strategies = make(map[string]func() BalancingStrategy, 0)
	strategies[StrategyRoundRobin] = func() BalancingStrategy {
		return &RoundRobin{Current: 0}
	}

	// TODO: Add other load balancing strategies here
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	//return (sl.current + 1) % uint32(len(sl.Servers))

	nxt := atomic.AddUint32(&r.Current, uint32(1))
	lenS := uint32(len(servers))

	// wrap it around whatever number of services we have
	//if nxt >= lenS {
	//	nxt -= lenS
	//}

	picked := servers[nxt%lenS]

	log.Infof("Strategy picked server '%s'", picked.Url.Host)

	return picked, nil
}

type WeightedRoundRobin struct {
	mu sync.Mutex

	/* This is making the assumption that the server list coming through the Next method, won't change between successive calls.
	Changing the server list would cause this strategy to break, panic or not route properly*/
	count   []int
	current int
}

func (w *WeightedRoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.count == nil {
		w.count = make([]int, len(servers))
		w.current = 0
	}

	capacity := servers[w.current].GetMetaOrDefault("weight", 1)

	if w.count[w.current] <= capacity {
		w.count[w.current]++

		return servers[w.current], nil
	}

	// server is at it's limit, reset the current server count and move on to the next server
	w.count[w.current] = 0
	w.current = (w.current + 1) % len(servers)

	return servers[w.current], nil
}

/*
LoadStrategy will try and resolve the balancing strategy based on the name and will default to a 'StrategyRoundRobin' one if
no strategy matched.
*/
func LoadStrategy(name string) BalancingStrategy {
	st, ok := strategies[name]
	if !ok {
		return strategies[StrategyRoundRobin]()
	}

	return st()
}
