package prometheus

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	localRegistry   = make(map[string]prometheus.Counter)
	localRegistryMu sync.RWMutex
)

// registerOnce registeres a counter with prometheus.Registerer if not registered yet.
func registerOnce(r prometheus.Registerer, o prometheus.CounterOpts) (prometheus.Counter, error) {
	id := descId(prometheus.Opts(o))

	// just return if already registered
	// trying read-lock first, optimizing for reads
	localRegistryMu.RLock()
	if c, ok := localRegistry[id]; ok {
		localRegistryMu.RUnlock()
		return c, nil
	}
	localRegistryMu.RUnlock()

	// not yet registered, we need exclusive lock
	localRegistryMu.Lock()
	if c, ok := localRegistry[id]; ok {
		// got registered in the meanwhile
		localRegistryMu.Unlock()
		return c, nil
	}

	// we hold the lock ...

	// create new counter
	c := prometheus.NewCounter(o)

	// register collector
	if err := r.Register(c); err != nil {
		localRegistryMu.Unlock()
		return nil, err
	}

	localRegistry[id] = c
	localRegistryMu.Unlock()
	return c, nil
}

// descId returns unique id for metric, ignoring labels because we don't use them here
func descId(c prometheus.Opts) string {
	return c.Namespace + "_" + c.Subsystem + "_" + c.Name
}
