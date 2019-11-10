package googleStackdriver

import (
	"io/ioutil"
	"os"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func TestLog(t *testing.T) {
	c := NewConfig()
	c.ClientOptions = []option.ClientOption{
		option.WithCredentialsJSON(readFile(t, ".credentials.json"))}
	c.LogName = os.Getenv("GOOGLE_PROJECT")
	c.LogID = "my-log"

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

	logger.Error("Hello world", Trace("trace123"))
}

func readFile(t *testing.T, file string) []byte {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
