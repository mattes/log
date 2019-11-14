package googleStackdriver

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"cloud.google.com/go/logging"
	"github.com/mattes/errorstats"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
	"google.golang.org/grpc/status"
)

// TODO make this configurable
var severity = map[zapcore.Level]logging.Severity{
	zapcore.DebugLevel:  logging.Debug,
	zapcore.InfoLevel:   logging.Info,
	zapcore.WarnLevel:   logging.Warning,
	zapcore.ErrorLevel:  logging.Error,
	zapcore.DPanicLevel: logging.Critical,
	zapcore.PanicLevel:  logging.Critical,
	zapcore.FatalLevel:  logging.Critical,
}

const (
	requestFieldKey                               = "_google_stackdriver_request"
	requestSizeFieldKey                           = "_google_stackdriver_request_size"
	requestStatusFieldKey                         = "_google_stackdriver_request_status"
	requestResponseSizeFieldKey                   = "_google_stackdriver_request_response_size"
	requestLatencyFieldKey                        = "_google_stackdriver_request_latency"
	requestLocalIPFieldKey                        = "_google_stackdriver_request_local_ip"
	requestRemoteIPFieldKey                       = "_google_stackdriver_request_remote_ip"
	requestCacheHitFieldKey                       = "_google_stackdriver_request_cache_hit"
	requestCacheValidatedWithOriginServerFieldKey = "_google_stackdriver_request_cache_validated_with_origin_server"
	traceFieldKey                                 = "_google_stackdriver_trace"
	traceSampledFieldKey                          = "_google_stackdriver_trace_sampled"
	spanIDFieldKey                                = "_google_stackdriver_span_id"
)

func Request(r *http.Request) zapcore.Field {
	return customField(requestFieldKey, r)
}

func RequestSize(size int64) zapcore.Field {
	return customField(requestSizeFieldKey, size)
}

func RequestStatus(status int) zapcore.Field {
	return customField(requestStatusFieldKey, status)
}

func RequestResponseSize(size int64) zapcore.Field {
	return customField(requestResponseSizeFieldKey, size)
}

func RequestLatency(latency time.Duration) zapcore.Field {
	return customField(requestLatencyFieldKey, latency)
}

func RequestLocalIP(ip string) zapcore.Field {
	return customField(requestLocalIPFieldKey, ip)
}

func RequestRemoteIP(ip string) zapcore.Field {
	return customField(requestRemoteIPFieldKey, ip)
}

func RequestCacheHit(hit bool) zapcore.Field {
	return customField(requestCacheHitFieldKey, hit)
}

func RequestCacheValidatedWithOriginServer(validated bool) zapcore.Field {
	return customField(requestCacheValidatedWithOriginServerFieldKey, validated)
}

func Trace(trace string) zapcore.Field {
	return customField(traceFieldKey, trace)
}

func TraceSampled(sampled bool) zapcore.Field {
	return customField(traceSampledFieldKey, sampled)
}

func SpanID(id string) zapcore.Field {
	return customField(spanIDFieldKey, id)
}

type Config struct {
	// Level is the minimum enabled logging level.
	// By default, only >= info levels are logged.
	Level zap.AtomicLevel

	// LogName can be either a project id or:
	// projects/PROJECT_ID
	// folders/FOLDER_ID
	// billingAccounts/ACCOUNT_ID
	// organizations/ORG_ID
	//
	// If empty, project id will be used from metadata server.
	LogName string

	// LogID must be less than 512 characters long and can only
	// include the following characters: upper and lower case alphanumeric
	// characters: [A-Za-z0-9]; and punctuation characters: forward-slash,
	// underscore, hyphen, and period.
	//
	// Example LogIDs: https://cloud.google.com/logging/docs/agent/default-logs#custom_types
	LogID string

	// BufferedByteLimit is the maximum number of bytes that the Logger will keep
	// in memory before returning ErrOverflow.
	BufferedByteLimit int

	// ConcurrentWriteLimit determines how many goroutines will send log entries
	// to the underlying service. Set ConcurrentWriteLimit to a higher value to
	// increase throughput.
	ConcurrentWriteLimit int

	// DelayThreshold is the maximum amount of time that an entry should remain
	// buffered in memory before a call to the logging service is triggered.
	// Larger values of DelayThreshold will generally result in fewer calls to
	// the logging service, while increasing the risk that log entries will be
	// lost if the process crashes.
	DelayThreshold time.Duration

	// EntryByteLimit is the maximum number of bytes of entries that will be sent
	// in a single call to the logging service. If EntryByteLimit is smaller than
	// EntryByteThreshold, the latter has no effect. The default is zero, meaning
	// there is no limit.
	EntryByteLimit int

	// EntryByteThreshold is the maximum number of bytes of entries that will be
	// buffered in memory before a call to the logging service is triggered.
	// See EntryCountThreshold for a discussion of the tradeoffs involved in
	// setting this option.
	EntryByteThreshold int

	// EntryCountThreshold is the maximum number of entries that will be buffered
	// in memory before a call to the logging service is triggered. Larger values
	// will generally result in fewer calls to the logging service, while increasing
	// both memory consumption and the risk that log entries will be lost if the
	// process crashes.
	EntryCountThreshold int

	// MonitoredResource sets the monitored resource associated with all log entries
	// written from a Logger. If not provided, the resource is automatically detected
	// based on the running environment (on GCE and GAE Standard only).
	// It translates to https://godoc.org/google.golang.org/genproto/googleapis/api/monitoredres#MonitoredResource
	MonitoredResourceType   string
	MonitoredResourceLabels map[string]string

	// Options for logging client, i.e.
	// option.WithCredentialsFile("credentials.json")
	ClientOptions []option.ClientOption
}

func NewConfig() Config {
	return Config{
		Level: zap.NewAtomicLevelAt(zap.InfoLevel),

		BufferedByteLimit:    logging.DefaultBufferedByteLimit,
		ConcurrentWriteLimit: runtime.NumCPU() * 4,
		DelayThreshold:       logging.DefaultDelayThreshold,
		EntryByteLimit:       0,
		EntryByteThreshold:   logging.DefaultEntryByteThreshold,
		EntryCountThreshold:  logging.DefaultEntryCountThreshold,
	}
}

func (cfg Config) Build() (zapcore.Core, error) {
	if cfg.LogID == "" {
		return nil, fmt.Errorf("missing LogID")
	}

	// use project id if no log name is given
	if cfg.LogName == "" {
		projectId, err := defaultProject()
		if err != nil {
			return nil, err
		}
		cfg.LogName = projectId
	}

	c := &core{}
	c.errs = errorstats.New()

	c.errs.SetEncoder(status.Status{}, func(v interface{}) string {
		x := v.(status.Status)
		return "grpc/status." + x.Code().String()
	})

	c.LevelEnabler = cfg.Level

	// create a console encoder to be used to marshal fields into json
	// for error reporting message
	c.fieldsEnc = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "",
		LevelKey:       "",
		NameKey:        "",
		CallerKey:      "",
		MessageKey:     "",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// create new logging client
	client, err := newClient(cfg.LogName, c.errs, cfg.ClientOptions...)
	if err != nil {
		return nil, err
	}
	c.client = client

	// create new logger
	loggerOpts := []logging.LoggerOption{
		logging.BufferedByteLimit(cfg.BufferedByteLimit),
		logging.ConcurrentWriteLimit(cfg.ConcurrentWriteLimit),
		logging.DelayThreshold(cfg.DelayThreshold),
		logging.EntryByteLimit(cfg.EntryByteLimit),
		logging.EntryByteThreshold(cfg.EntryByteThreshold),
		logging.EntryCountThreshold(cfg.EntryCountThreshold),
	}

	if cfg.MonitoredResourceType != "" || len(cfg.MonitoredResourceLabels) > 0 {
		loggerOpts = append(loggerOpts, logging.CommonResource(
			&monitoredres.MonitoredResource{
				Type:   cfg.MonitoredResourceType,
				Labels: cfg.MonitoredResourceLabels,
			}))
	}

	logger, err := newLogger(client, cfg.LogID, loggerOpts...)
	if err != nil {
		return nil, err
	}
	c.logger = logger

	return c, nil
}

type core struct {
	zapcore.LevelEnabler

	client    *logging.Client
	logger    *logging.Logger
	fieldsEnc zapcore.Encoder
	errs      *errorstats.Stats // internal errors
}

func (c *core) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()

	for i := range fields {
		fields[i].AddTo(clone.fieldsEnc)
	}

	return clone
}

func (c *core) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}

	return checkedEntry
}

func (c *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	e := logging.Entry{}
	e.Labels = make(map[string]string)

	// TODO should we generate a uuid for the e.InsertID here?

	e.Timestamp = entry.Time

	// set severity
	if s, ok := severity[entry.Level]; ok {
		e.Severity = s
	} else {
		e.Severity = logging.Default
	}

	e.HTTPRequest = &logging.HTTPRequest{Request: &http.Request{URL: &url.URL{}}}
	httpRequestSet := false

	for _, f := range fields {
		switch f.Key {

		case requestFieldKey:
			e.HTTPRequest.Request = f.Interface.(*http.Request)
			httpRequestSet = true

		case requestSizeFieldKey:
			e.HTTPRequest.RequestSize = f.Interface.(int64)
			httpRequestSet = true

		case requestStatusFieldKey:
			e.HTTPRequest.Status = f.Interface.(int)
			httpRequestSet = true

		case requestResponseSizeFieldKey:
			e.HTTPRequest.ResponseSize = f.Interface.(int64)
			httpRequestSet = true

		case requestLatencyFieldKey:
			e.HTTPRequest.Latency = f.Interface.(time.Duration)
			httpRequestSet = true

		case requestLocalIPFieldKey:
			e.HTTPRequest.LocalIP = f.Interface.(string)
			httpRequestSet = true

		case requestRemoteIPFieldKey:
			e.HTTPRequest.RemoteIP = f.Interface.(string)
			httpRequestSet = true

		case requestCacheHitFieldKey:
			e.HTTPRequest.CacheHit = f.Interface.(bool)
			httpRequestSet = true

		case requestCacheValidatedWithOriginServerFieldKey:
			e.HTTPRequest.CacheValidatedWithOriginServer = f.Interface.(bool)
			httpRequestSet = true

		case traceFieldKey:
			e.Trace = f.Interface.(string)

		case traceSampledFieldKey:
			e.TraceSampled = f.Interface.(bool)

		case spanIDFieldKey:
			e.SpanID = f.Interface.(string)
		}
	}

	if !httpRequestSet {
		e.HTTPRequest = nil
	}

	// marshal fields into json for human output
	// TODO should we use e.Labels instead?
	buf, err := c.fieldsEnc.EncodeEntry(zapcore.Entry{}, fields)
	if err != nil {
		return err
	}
	fieldsStr := buf.String()

	errorStr := entry.Message
	if len(fieldsStr) > 0 {
		errorStr += " " + fieldsStr
	}

	e.Payload = errorStr

	if entry.Caller.Defined {
		e.SourceLocation = &logpb.LogEntrySourceLocation{
			File:     entry.Caller.File,
			Line:     int64(entry.Caller.Line),
			Function: funcNameForPC(entry.Caller.PC),
		}
	}

	if entry.LoggerName != "" {
		e.Labels["logger"] = entry.LoggerName
	}

	// log message
	c.logger.Log(e)

	if entry.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output
		// errors, pending a clean solution to issue #370.
		// https://github.com/uber-go/zap/issues/370
		c.Sync()
	}

	// see if any errors have been logged in the meanwhile,
	// if yes, return them here.
	return c.errs.ErrAndReset()
}

func (c *core) Sync() error {
	return multierr.Combine(
		c.logger.Flush(),
		c.errs.ErrAndReset())
}

func (c *core) clone() *core {
	return &core{
		LevelEnabler: c.LevelEnabler,
		client:       c.client,
		logger:       c.logger,
		fieldsEnc:    c.fieldsEnc.Clone(),
		errs:         c.errs,
	}
}

func funcNameForPC(pc uintptr) string {
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}

	return f.Name()
}

// customField returns a zapcore.Field that is skipped by other cores,
// and only has special meaning to his core.
func customField(key string, v interface{}) zapcore.Field {
	return zapcore.Field{
		Key:       key,
		Type:      zapcore.SkipType,
		Interface: v,
	}
}
