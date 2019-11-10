package googleStackdriver

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
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
	return zap.Any(requestFieldKey, r)
}

func RequestSize(size int64) zapcore.Field {
	return zap.Int64(requestSizeFieldKey, size)
}

func RequestStatus(status int) zapcore.Field {
	return zap.Int(requestStatusFieldKey, status)
}

func RequestResponseSize(size int64) zapcore.Field {
	return zap.Int64(requestResponseSizeFieldKey, size)
}

func RequestLatency(latency time.Duration) zapcore.Field {
	return zap.Duration(requestLatencyFieldKey, latency)
}

func RequestLocalIP(ip string) zapcore.Field {
	return zap.String(requestLocalIPFieldKey, ip)
}

func RequestRemoteIP(ip string) zapcore.Field {
	return zap.String(requestRemoteIPFieldKey, ip)
}

func RequestCacheHit(hit bool) zapcore.Field {
	return zap.Bool(requestCacheHitFieldKey, hit)
}

func RequestCacheValidatedWithOriginServer(validated bool) zapcore.Field {
	return zap.Bool(requestCacheValidatedWithOriginServerFieldKey, validated)
}

func Trace(trace string) zapcore.Field {
	return zap.String(traceFieldKey, trace)
}

func TraceSampled(sampled bool) zapcore.Field {
	return zap.Bool(traceSampledFieldKey, sampled)
}

func SpanID(id string) zapcore.Field {
	return zap.String(spanIDFieldKey, id)
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
	var f *zapcore.Field
	httpRequestSet := false

	fields, f = filterFields(requestFieldKey, fields)
	if f != nil {
		if v, ok := f.Interface.(*http.Request); ok {
			e.HTTPRequest.Request = v
			httpRequestSet = true
		} else {
			panic(fmt.Sprintf("expected *http.Request for field with key '%v'", requestFieldKey))
		}
	}

	fields, f = filterFields(requestSizeFieldKey, fields)
	if f != nil {
		e.HTTPRequest.RequestSize = f.Integer
		httpRequestSet = true
	}

	fields, f = filterFields(requestStatusFieldKey, fields)
	if f != nil {
		e.HTTPRequest.Status = int(f.Integer)
		httpRequestSet = true
	}

	fields, f = filterFields(requestResponseSizeFieldKey, fields)
	if f != nil {
		e.HTTPRequest.ResponseSize = f.Integer
		httpRequestSet = true
	}

	fields, f = filterFields(requestLatencyFieldKey, fields)
	if f != nil {
		e.HTTPRequest.Latency = time.Duration(f.Integer)
		httpRequestSet = true
	}

	fields, f = filterFields(requestLocalIPFieldKey, fields)
	if f != nil {
		e.HTTPRequest.LocalIP = f.String
		httpRequestSet = true
	}

	fields, f = filterFields(requestRemoteIPFieldKey, fields)
	if f != nil {
		e.HTTPRequest.RemoteIP = f.String
		httpRequestSet = true
	}

	fields, f = filterFields(requestCacheHitFieldKey, fields)
	if f != nil {
		if f.Type != zapcore.BoolType {
			panic(fmt.Sprintf("expected bool for field with key '%v'", requestCacheHitFieldKey))
		}
		httpRequestSet = true
		if f.Integer == 1 {
			e.HTTPRequest.CacheHit = true
		}
	}

	fields, f = filterFields(requestCacheValidatedWithOriginServerFieldKey, fields)
	if f != nil {
		if f.Type != zapcore.BoolType {
			panic(fmt.Sprintf("expected bool for field with key '%v'", requestCacheValidatedWithOriginServerFieldKey))
		}
		httpRequestSet = true
		if f.Integer == 1 {
			e.HTTPRequest.CacheValidatedWithOriginServer = true
		}
	}

	if !httpRequestSet {
		e.HTTPRequest = nil
	}

	fields, f = filterFields(traceFieldKey, fields)
	if f != nil {
		e.Trace = f.String
	}

	fields, f = filterFields(traceSampledFieldKey, fields)
	if f != nil {
		if f.Type != zapcore.BoolType {
			panic(fmt.Sprintf("expected bool for field with key '%v'", traceSampledFieldKey))
		}
		if f.Integer == 1 {
			e.TraceSampled = true
		}
	}

	fields, f = filterFields(spanIDFieldKey, fields)
	if f != nil {
		e.SpanID = f.String
	}

	// marshal fields into json for human output
	// TODO should we use e.Labels instead?
	fields = filterSpecialFields(fields)
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

func filterFields(key string, fields []zapcore.Field) ([]zapcore.Field, *zapcore.Field) {
	var f zapcore.Field

	n := fields[:0]
	for _, x := range fields {
		if x.Key == key {
			f = x
		} else {
			n = append(n, x)
		}
	}

	if f.Key == "" {
		return n, nil
	}

	return n, &f
}

func filterSpecialFields(fields []zapcore.Field) []zapcore.Field {
	n := fields[:0]
	for _, x := range fields {
		if !strings.HasPrefix(x.Key, "_") {
			n = append(n, x)
		}
	}
	return n
}
