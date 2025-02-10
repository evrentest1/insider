package httpkit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const (
	timeout          = 10 * time.Second
	retryWaitMax     = 5 * time.Second
	retryWaitMinimum = 3 * time.Second
	retryMax         = 2
	retryAfterMaxSec = 10
)

var retryConfig = RetryableClientConfig{
	Timeout:          timeout,
	RetryWaitMax:     retryWaitMax,
	RetryWaitMin:     retryWaitMinimum,
	RetryMax:         retryMax,
	RetryAfterMaxSec: retryAfterMaxSec,
}

type (
	RetryableClientConfig struct {
		Timeout          time.Duration
		RetryWaitMax     time.Duration
		RetryWaitMin     time.Duration
		RetryMax         int
		RetryAfterMaxSec int64
	}

	retryableClientServerError struct {
		StatusCode int
		Status     string
	}

	retryableClient struct {
		retryablehttp.Client
	}
)

func newRetryableClient(_ context.Context, client *http.Client, log *logrus.Logger) *retryableClient {
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = retryConfig.Timeout
	rc := &retryableClient{
		Client: retryablehttp.Client{
			HTTPClient: client,
			Logger:     log,
		},
	}
	rc.CheckRetry = rc.timeoutRetryPolicy
	return rc
}

func (rc *retryableClient) config(conf *RetryableClientConfig) *retryableClient {
	if conf == nil || conf.Timeout == 0 {
		return rc
	}

	rc.HTTPClient.Timeout = conf.Timeout
	rc.RetryWaitMax = conf.RetryWaitMax
	rc.RetryWaitMin = conf.RetryWaitMin
	rc.RetryMax = conf.RetryMax
	return rc
}

func (rc *retryableClient) timeoutRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		var v *url.Error
		if errors.As(err, &v) {
			return false, v
		}

		return true, err
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusNotImplemented) {
		return true, &retryableClientServerError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	return false, nil
}

func (e *retryableClientServerError) Error() string {
	return fmt.Sprintf("server error: %d - %s", e.StatusCode, e.Status)
}
