package slack

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLog(t *testing.T) {
	c := NewConfig()
	c.WebhookURL = os.Getenv("SLACK_URL")
	c.Channel = os.Getenv("SLACK_CHANNEL")

	core, err := c.Build()
	if err != nil {
		t.Fatal(err)
	}

	logger := zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	defer func() {
		if err := logger.Sync(); err != nil {
			t.Fatal(err)
		}
	}()

	logger.Error("Hello world")
}
