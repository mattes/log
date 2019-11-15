package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/zap"
)

func TestPrometheus(t *testing.T) {
	c := NewConfig()
	core, err := c.Build()
	if err != nil {
		t.Fatal(err)
	}

	logger := zap.New(core)
	defer func() {
		if err := logger.Sync(); err != nil {
			t.Fatal(err)
		}
	}()

	logger.Error("Something bad happened", Inc("something_bad"))
	logger.Error("Something bad happened", Inc("something_bad"))

	counter := testutil.ToFloat64(localRegistry[descId(prometheus.Opts{Name: "something_bad"})])
	if counter != 2 {
		t.Errorf("expected counter to be 2, got %v", counter)
	}
}
