package prometheus

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

const (
	increaseCounterFieldKey = "github.com/mattes/log/prometheus/increaseCounter"
)

// Inc returns a Field that creates a counter on the fly and increases it.
// An optional help text can be specified. If multiple help texts are given,
// they are concated into one string.
// If Config.UseMessageAsHelp is true and no help text is specified, the logged
// message is used as help text.
func Inc(name string, help ...string) zapcore.Field {
	o := prometheus.CounterOpts{}
	o.Name = name

	switch len(help) {
	case 0:
	// ignore

	case 1:
		o.Help = help[0]

	default:
		// just concat string
		o.Help = strings.Join(help, " ")
	}

	return customField(increaseCounterFieldKey, o)
}

type Config struct {
	Registerer prometheus.Registerer

	Namespace string
	Subsystem string

	// If Inc() is used without a help text and UseMessageAsHelp
	// is set to true (default false), the logged message will be used as help text.
	UseMessageAsHelp bool
}

func NewConfig() Config {
	return Config{
		Registerer: prometheus.DefaultRegisterer,
	}
}

func (cfg Config) Build() (zapcore.Core, error) {
	c := &core{
		Registerer:       cfg.Registerer,
		Namespace:        cfg.Namespace,
		Subsystem:        cfg.Subsystem,
		UseMessageAsHelp: cfg.UseMessageAsHelp,
	}

	return c, nil
}

type core struct {
	Registerer       prometheus.Registerer
	Namespace        string
	Subsystem        string
	UseMessageAsHelp bool
}

func (c *core) Enabled(zapcore.Level) bool {
	return true // always enabled, regardless of log level
}

func (c *core) With(fields []zapcore.Field) zapcore.Core {
	return c // we don't care about adding fields to this core
}

func (c *core) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, c)
}

func (c *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var o prometheus.CounterOpts
	found := false

	// find our key
	for _, f := range fields {
		if f.Key == increaseCounterFieldKey {
			o = f.Interface.(prometheus.CounterOpts)
			found = true
			break
		}
	}

	// key not found
	if !found {
		return nil
	}

	// if no help text was set, we can use the message as help text
	if c.UseMessageAsHelp && o.Help == "" {
		o.Help = entry.Message
	}

	// create counter and register once
	if counter, err := registerOnce(c.Registerer, o); err != nil {
		return err
	} else {
		counter.Inc()
	}

	return nil
}

func (c *core) Sync() error {
	return nil
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
