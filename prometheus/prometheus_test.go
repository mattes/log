package prometheus

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterOnce(t *testing.T) {
	localRegistry = make(map[string]prometheus.Counter)
	r := prometheus.NewRegistry()

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			o := prometheus.CounterOpts{Name: "id123"}
			if _, err := registerOnce(r, o); err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	f, err := r.Gather()
	if err != nil {
		t.Fatal(err)
	}

	if len(f) != 1 {
		t.Errorf("expected one registered collector, got %v", len(f))
	}

	if len(localRegistry) != 1 {
		t.Errorf("expected one registered collector in localRegistry, got %v", len(localRegistry))
	}
}

func TestRegisterWithDifferentHelp(t *testing.T) {
	// prometheus.Register will return an error if help text is different
	{
		r := prometheus.NewRegistry()

		c1 := prometheus.NewCounter(prometheus.CounterOpts{Name: "foo", Help: "bar"})
		if err := r.Register(c1); err != nil {
			t.Fatal(err)
		}

		c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "foo", Help: "rab"})
		if err := r.Register(c2); err == nil {
			t.Fatal("expected err")
		}
	}

	// registerOnce will work fine with different help texts
	{
		r := prometheus.NewRegistry()

		if _, err := registerOnce(r, prometheus.CounterOpts{Name: "foo", Help: "bar"}); err != nil {
			t.Fatal(err)
		}
		if _, err := registerOnce(r, prometheus.CounterOpts{Name: "foo", Help: "rab"}); err != nil {
			t.Fatal(err)
		}
	}
}
