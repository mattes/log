package slack

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/mattes/errorstats"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/support/bundler"
)

// TODO make this configurable
// https://material.io/design/color/#tools-for-picking-colors
var colors = map[zapcore.Level]string{
	zapcore.DebugLevel:  "#616161",
	zapcore.InfoLevel:   "#1976D2",
	zapcore.WarnLevel:   "#FBC02D",
	zapcore.ErrorLevel:  "#D32F2F",
	zapcore.DPanicLevel: "#7B1FA2",
	zapcore.PanicLevel:  "#7B1FA2",
	zapcore.FatalLevel:  "#7B1FA2",
}

type Config struct {
	// Level is the minimum enabled logging level.
	// By default, only >= errors are send to Slack.
	Level zap.AtomicLevel

	// WebhookURL is the Slack Webhook URL
	WebhookURL string

	// Slack channel
	Channel string

	// BatchDelayThreshold defines the time to wait before a batch is flushed after
	// a message is logged.
	BatchDelayThreshold time.Duration

	// BatchCountThreshold defines the maximum messages in a batch.
	BatchCountThreshold int

	// BatchHandlerLimit sets how many batches can be processed at the same time.
	BatchHandlerLimit int
}

func NewConfig() Config {
	return Config{
		Level:               zap.NewAtomicLevelAt(zap.ErrorLevel),
		BatchDelayThreshold: 1 * time.Second,
		BatchCountThreshold: 20, // if higher than 20, Slack will start hiding messages
		BatchHandlerLimit:   runtime.NumCPU() * 10,
	}
}

func (cfg Config) Build() (zapcore.Core, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("missing WebhookURL")
	}

	c := &core{}
	c.errs = errorstats.New()
	c.LevelEnabler = cfg.Level

	// create a console encoder to be used to marshal fields into json
	// for slack message
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

	// create bundler
	c.bundle = bundler.NewBundler(&slackAttachment{},
		func(entries interface{}) {
			bundleHandler(cfg.WebhookURL, cfg.Channel, c.errs, entries.([]*slackAttachment))
		})

	c.bundle.DelayThreshold = cfg.BatchDelayThreshold
	c.bundle.BundleCountThreshold = cfg.BatchCountThreshold
	c.bundle.HandlerLimit = cfg.BatchHandlerLimit

	return c, nil
}

// bundleHandler is called by bundler when batch is full
func bundleHandler(webhookURL, channel string, errs *errorstats.Stats, slackAttachments []*slackAttachment) {
	p := &slackPayload{
		Channel: channel,
	}

	// add messages from batch as slack attachments
	for _, a := range slackAttachments {
		p.Attachments = append(p.Attachments, a)
	}

	err := sendMessage(webhookURL, p)
	if err != nil {
		errs.Log(err)
	}
}

type core struct {
	zapcore.LevelEnabler

	fieldsEnc zapcore.Encoder
	bundle    *bundler.Bundler
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
	// marshal fields into json for human output
	fields = filterSpecialFields(fields)
	buf, err := c.fieldsEnc.EncodeEntry(zapcore.Entry{}, fields)
	if err != nil {
		return err
	}
	fieldsStr := buf.String()

	// TODO create some sort of slack message encoder to allow overwriting
	// how a slack message is formatted and styled

	a := &slackAttachment{}
	a.Fallback = entry.Level.CapitalString() + " " + entry.Message
	a.Timestamp = entry.Time.Unix()
	a.Title = entry.Message
	a.MarkdownIn = []string{"text"}
	a.Footer = entry.Level.CapitalString()

	// set text field
	if len(fieldsStr) > 0 {
		a.Text = fieldsStr
	}

	if len(entry.Stack) > 0 {
		a.Text += "\n```" + entry.Stack + "```"
	} else if entry.Caller.Defined {
		a.Text += "\n```" + entry.Caller.String() + "```"
	}

	if entry.LoggerName != "" {
		a.Text += "\n" + entry.LoggerName
		a.Fallback += " (" + entry.LoggerName + ")"
	}

	// set color
	if color, ok := colors[entry.Level]; ok {
		a.Color = color
	} else {
		a.Color = "#FFFFFF"
	}

	if err := c.bundle.Add(a, 1); err != nil {
		return err
	}

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
	c.bundle.Flush() // returns no error
	return c.errs.ErrAndReset()
}

func (c *core) clone() *core {
	return &core{
		LevelEnabler: c.LevelEnabler,
		fieldsEnc:    c.fieldsEnc.Clone(),
		bundle:       c.bundle,
		errs:         c.errs,
	}
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
