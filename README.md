# log [![GoDoc](https://godoc.org/github.com/mattes/log?status.svg)](https://godoc.org/github.com/mattes/log)

log acts as drop-in replacement for `std/log` and uses the power of
[Uber's Zap](https://github.com/uber-go/zap) logging library internally. 

It implements the following Zap cores:

  * [Google Cloud Error Reporting](/googleErrorReporting)
  * [Google Cloud Stackdriver Logging](/googleStackdriver)
  * [Slack](/slack)
  * [Prometheus](/prometheus)


## Usage

A __development__ setup is used by default. Similar to `std/log`.
All logs are written to stderr.

```go
import "github.com/mattes/log"

func main() {
  defer log.Sync()

  log.Info("Hello world")
}
```

In __production__ the setup depends on where you want to ship your logs to.
Here is an example that ships all logs to Google Stackdriver, as well
as Google Error Reporting.

```go
import (
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
  "github.com/mattes/log"
  gerr "github.com/mattes/log/googleErrorReporting"
  gsdr "github.com/mattes/log/googleStackdriver"

)

func init() {
  cores := []zapcore.Core{}

  {
    // Stackdriver core
    c := gsdr.NewConfig()
    c.LogID = "my-service.v2"
    core, err := c.Build()
    cores = append(cores, core)
  }

  {
    // Error reporting core
    c := gerr.NewConfig()
    c.ServiceName = "my-service"
    c.ServiceVersion = "v2"
    core, err := c.Build()
    cores = append(cores, core)
  }

  // Build Zap logger with options
  logger := zap.New(zapcore.NewTee(cores...)).WithOptions(
    zap.AddCaller(),
    zap.AddCallerSkip(1),
    zap.AddStacktrace(zapcore.ErrorLevel),
    log.ErrorOutput("stderr"), // for internal errors
    log.Sampling(100, 100),
  )

  // Set global logger
  log.Use(logger)
}

func main() {
  defer log.Sync()
  defer log.CapturePanic() // optional, it logs unhandled panics before crashing

  log.Error("Hello Production!")
}
```

## Changing log level

Update the config to use a reference of [zap#AtomicLevel](https://godoc.org/go.uber.org/zap#NewAtomicLevel)
that you control. It can serve as [HTTP handler](https://godoc.org/go.uber.org/zap#AtomicLevel.ServeHTTP), too.


## Replacing logger in third-party lib

Sometimes third-party libraries log on their own with no way of disabling it.
With the help of Go modules we can replace third-party logging libraries 
and instruct them to log through our logging infrastructure.

  * [golang/glog](/glog)


## Testing

Some tests require credentials and configration that can't be commited to the repo.
I recommend putting a `.env` file in each directory with contents like:

```
export SLACK_URL='https://hooks.slack.com/services/xxx'
```

Then run test with `source .env && go test -v`

