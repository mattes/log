package googleErrorReporting

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func TestLog(t *testing.T) {
	c := NewConfig()
	c.Project = os.Getenv("GOOGLE_PROJECT")
	c.ClientOptions = []option.ClientOption{
		option.WithCredentialsJSON(readFile(t, ".credentials.json"))}
	c.ServiceName = "my-service"
	c.ServiceVersion = "v2"

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

	fakeRequest := &http.Request{
		Method:     "GET",
		Host:       "example.com",
		RequestURI: "/",
	}

	logger.Sugar().Errorw("Hello world", User("user123"), Request(fakeRequest))
}

func readFile(t *testing.T, file string) []byte {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
