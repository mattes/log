package googleStackdriver

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/mattes/errorstats"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
)

func defaultProject() (string, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	m := metadata.NewClient(httpClient)
	return m.ProjectID()
}

func newClient(logName string, errs *errorstats.Stats, opts ...option.ClientOption) (*logging.Client, error) {
	c, err := logging.NewClient(context.Background(), logName, opts...)
	if err != nil {
		return nil, err
	}

	c.OnError = func(err error) {
		if s, ok := status.FromError(err); ok {
			errs.Log(s)
		} else {
			errs.Log(err)
		}
	}

	return c, err
}

func newLogger(c *logging.Client, logID string, opts ...logging.LoggerOption) (*logging.Logger, error) {
	if !validLogID(logID) {
		return nil, fmt.Errorf("invalid logID")
	}

	return c.Logger(logID, opts...), nil
}

// A log ID must be less than 512 characters long and can only
// include the following characters: upper and lower case alphanumeric
// characters: [A-Za-z0-9]; and punctuation characters: forward-slash,
// underscore, hyphen, and period.
var logIDRe = regexp.MustCompile(`^[A-Za-z0-9\/_\-\.]{1,512}$`)

func validLogID(logID string) bool {
	return logIDRe.MatchString(logID)
}
