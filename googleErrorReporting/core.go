package googleErrorReporting

import (
	"bytes"
	"errors"
	"net/http"
	"runtime"

	"cloud.google.com/go/errorreporting"
	"github.com/mattes/errorstats"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
)

const (
	userFieldKey    = "github.com/mattes/log/googleErrorReporting/user"
	requestFieldKey = "github.com/mattes/log/googleErrorReporting/request"
)

// User adds user id to log
func User(id string) zapcore.Field {
	return customField(userFieldKey, id)
}

// Request adds http request to log
func Request(r *http.Request) zapcore.Field {
	return customField(requestFieldKey, r)
}

type Config struct {
	// Level is the minimum enabled logging level.
	// By default, only >= errors are send to Slack.
	Level zap.AtomicLevel

	// Project sets the Google Cloud Project. If empty the GCE metadata server is asked for it.
	Project string

	// ServiceName identifies the running program and is included in the error reports.
	ServiceName string

	// ServiceVersion identifies the version of the running program and is
	// included in the error reports.
	ServiceVersion string

	// Options for error reporting client, i.e.
	// option.WithCredentialsFile("credentials.json")
	ClientOptions []option.ClientOption
}

func NewConfig() Config {
	return Config{
		Level: zap.NewAtomicLevelAt(zap.ErrorLevel),
	}
}

func (cfg Config) Build() (zapcore.Core, error) {
	// set project to default if empty
	if cfg.Project == "" {
		projectId, err := defaultProject()
		if err != nil {
			return nil, err
		}
		cfg.Project = projectId
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

	// create new error reporting client
	client, err := newClient(cfg.Project, cfg.ServiceName, cfg.ServiceVersion, c.errs, cfg.ClientOptions...)
	if err != nil {
		return nil, err
	}
	c.client = client

	return c, nil
}

type core struct {
	zapcore.LevelEnabler

	client    *errorreporting.Client
	fieldsEnc zapcore.Encoder
	errs      *errorstats.Stats
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
	r := errorreporting.Entry{}

	for _, f := range fields {
		switch f.Key {

		case requestFieldKey:
			r.Req = f.Interface.(*http.Request)

		case userFieldKey:
			r.User = f.Interface.(string)
		}
	}

	// marshal fields into json for human output
	buf, err := c.fieldsEnc.EncodeEntry(zapcore.Entry{}, fields)
	if err != nil {
		return err
	}
	fieldsStr := buf.String()

	// TODO create some sort of error reporting entry encoder to allow overwriting
	// how a report is formatted

	errorStr := entry.Message
	if len(fieldsStr) > 0 {
		errorStr += "\n" + fieldsStr
	}

	if entry.LoggerName != "" {
		errorStr += "\n(" + entry.LoggerName + ""
	}

	r.Error = errors.New(errorStr)

	// Add stacktrace.
	// Ignore entry.Stack which will be empty anyway.
	// Also, it appears that entry.Stack doesn't conform with
	// https://cloud.google.com/error-reporting/reference/rest/v1beta1/projects.events/report#ReportedErrorEvent.message
	// Limit the stack trace to 16k.
	var sbuf [16 * 1024]byte
	r.Stack = trimStack(sbuf[0:runtime.Stack(sbuf[:], false)])

	// send error report
	c.client.Report(r)

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
	c.client.Flush() // returns no errors
	return c.errs.ErrAndReset()
}

func (c *core) clone() *core {
	return &core{
		LevelEnabler: c.LevelEnabler,
		client:       c.client,
		fieldsEnc:    c.fieldsEnc.Clone(),
		errs:         c.errs,
	}
}

// trimStack removes callers from stack
// modified from here: https://github.com/googleapis/google-cloud-go/blob/master/errorreporting/errors.go
func trimStack(s []byte) []byte {
	// where does the first line end?
	firstLine := bytes.IndexByte(s, '\n')
	if firstLine == -1 {
		return s
	}

	// stack is stack without first line
	stack := s[firstLine:]

	// find last go.uber.org/zap Logger or SugaredLogger
	findLine := bytes.LastIndex(stack, []byte("\ngo.uber.org/zap.(*Logger)."))
	if findLine == -1 {
		findLine = bytes.LastIndex(stack, []byte("\ngo.uber.org/zap.(*SugaredLogger)"))
	}
	if findLine == -1 {
		return s
	}

	// stack starts where we found the line from above
	stack = stack[findLine:]

	// skip the next 3 lines
	for i := 0; i < 3; i++ {
		nextLine := bytes.IndexByte(stack, '\n')
		if nextLine == -1 {
			return s
		}
		stack = stack[nextLine+1:]
	}

	// merge first line and remaining stack
	return append(s[:firstLine+1], stack...)
}

// customField returns a zapcore.Field that is skipped by other cores,
// and only has special meaning to this core.
func customField(key string, v interface{}) zapcore.Field {
	return zapcore.Field{
		Key:       key,
		Type:      zapcore.SkipType,
		Interface: v,
	}
}
