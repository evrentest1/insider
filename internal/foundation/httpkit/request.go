package httpkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

var (
	ErrHTTPStatus          = errors.New("HTTP status error")
	ErrHTTPResponse        = errors.New("HTTP response error")
	ErrNilStruct           = errors.New("nil struct")
	ErrHTTPResponseMarshal = errors.New("failed to marshal http response")
	ErrHTTPRequestCreate   = errors.New("failed to create http request")
)

type RequesterService interface {
	Post(ctx context.Context, url string, headers map[string]string, body, resp any) (ResponseCode, http.Header, RawResponse, error)
}

type HTTPClientService interface {
	HTTPClient() *http.Client
}

type Requester struct {
	Headers     *map[string]string
	RetryConfig *RetryableClientConfig

	httpClient *http.Client
	log        *logrus.Logger
}

type (
	RawResponse  []byte
	ResponseCode int
)

func NewRequester(log *logrus.Logger) *Requester {
	return &Requester{
		httpClient: &http.Client{},
		log:        log,
	}
}

func (r *Requester) Post(ctx context.Context, url string, headers map[string]string, body, resp any) (ResponseCode, http.Header, RawResponse, error) {
	if resp == nil {
		return 0, nil, nil, fmt.Errorf("%w: response cant bind to a nil struct", ErrNilStruct)
	}

	req, err := r.createRequest(ctx, "POST", url, body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%w: %w", ErrHTTPRequestCreate, err)
	}

	r.fillHeaders(req, headers)

	code, header, data, err := r.getResponse(ctx, req)
	if err != nil {
		return code, header, data, fmt.Errorf("%w: %w", ErrHTTPResponse, err)
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return code, header, data, fmt.Errorf("%w: %w", ErrHTTPResponseMarshal, err)
	}

	return code, header, data, nil
}

func (r *Requester) createRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	var buff io.Reader
	if body != nil {
		bb, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buff = bytes.NewBuffer(bb)
	}

	return http.NewRequestWithContext(ctx, method, url, buff)
}

func (r *Requester) fillHeaders(req *http.Request, header map[string]string) {
	for k, v := range header {
		req.Header.Set(k, v)
	}
}

func (r *Requester) getResponse(ctx context.Context, req *http.Request) (ResponseCode, http.Header, []byte, error) {
	req.Header.Set("User-Agent", "insider")
	req.Header.Set("Content-Type", "application/json")

	client := newRetryableClient(ctx, r.httpClient, r.log).config(r.RetryConfig)
	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return 0, nil, nil, err
	}
	retryReq = retryReq.WithContext(ctx)

	res, err := client.Do(retryReq)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		var serverError *retryableClientServerError
		if errors.As(err, &serverError) {
			return ResponseCode(serverError.StatusCode), nil, nil, err
		}
		return 0, nil, nil, err
	}
	if res == nil {
		return 0, nil, nil, fmt.Errorf("%w: response is nil", ErrHTTPResponse)
	}

	status := res.StatusCode
	if status == http.StatusNoContent {
		return ResponseCode(status), res.Header, nil, nil
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return ResponseCode(status), res.Header, data, err
	}

	if status >= http.StatusMultipleChoices {
		return ResponseCode(status), res.Header, data, fmt.Errorf("%w: code: %d, message: %s", ErrHTTPStatus, status, string(data))
	}

	return ResponseCode(status), res.Header, data, nil
}
