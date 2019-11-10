package googleErrorReporting

import (
	"context"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/errorreporting"
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

func newClient(project, serviceName, serviceVersion string, errs *errorstats.Stats, opts ...option.ClientOption) (*errorreporting.Client, error) {
	c, err := errorreporting.NewClient(
		context.Background(),
		project,
		errorreporting.Config{
			ServiceName:    serviceName,
			ServiceVersion: serviceVersion,
			OnError: func(err error) {
				if s, ok := status.FromError(err); ok {
					errs.Log(s)
				} else {
					errs.Log(err)
				}
			},
		},
		opts...)

	return c, err
}
