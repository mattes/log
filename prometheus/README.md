# Prometheus [![GoDoc](https://godoc.org/github.com/mattes/log/prometheus?status.svg)](https://godoc.org/github.com/mattes/log/prometheus)

This package implements a Prometheus core using the [prometheus/client_golang](https://github.com/prometheus/client_golang) lib.
It conveniently allows to increase a counter for a logged message. See example below.


## Usage

```go
import (
  "go.uber.org/zap"
  prom "github.com/mattes/log/prometheus"
)

c := prom.NewConfig()

core, err := c.Build()
if err != nil {
  panic(err)
}

logger := zap.New(core)
defer logger.Sync()

logger.Error("Something bad happened", prom.Inc("something_bad"))
```

## Notes

* Implications of changing help texts, see [Stackoverflow](https://stackoverflow.com/questions/58853409/implications-of-a-prometheus-metric-with-different-help-texts)
* [Metric name docs](https://prometheus.io/docs/practices/naming/#metric-names) 

