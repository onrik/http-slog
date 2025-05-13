package httpslog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Slog interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
}

type Config struct {
	Transport http.RoundTripper
	Logger    Slog
}

type transport struct {
	rt   http.RoundTripper
	slog Slog
}

// NewTransport return http.RoundTripper with logging requests and responses
func New(config *Config) http.RoundTripper {
	if config == nil {
		config = &Config{}
	}

	if config.Transport == nil {
		config.Transport = http.DefaultTransport
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return &transport{
		rt:   config.Transport,
		slog: config.Logger,
	}
}

func (t *transport) logRequest(request *http.Request) error {
	body, err := CopyRequestBody(request)
	if err != nil {
		return fmt.Errorf("copy request body error: %w", err)
	}

	fields := []any{
		"headers", request.Header,
		"method", request.Method,
		"url", request.URL.String(),
		"body", string(body),
	}

	if len(request.URL.Query()) > 0 {
		fields = append(fields, "query", request.URL.Query())
	}

	message := fmt.Sprintf("Request-> %s %s", request.Method, request.URL.Path)
	t.slog.InfoContext(request.Context(), message, fields...)

	return nil
}

func (t *transport) logResponse(request *http.Request, response *http.Response, responseErr error, latency time.Duration) error {
	message := fmt.Sprintf("Response<- %s %s", request.Method, request.URL.Path)
	fields := []any{
		"latency", latency.String(),
		"method", request.Method,
		"url", request.URL.String(),
	}
	if responseErr != nil {
		fields = append(fields, "error", responseErr)
		t.slog.ErrorContext(request.Context(), message, fields...)
		return responseErr
	}

	body, err := CopyResponseBody(response)
	if err != nil {
		return fmt.Errorf("copy response body error: %w", err)
	}

	fields = append(fields, "status", response.StatusCode, "headers", response.Header, "body", string(body))
	if response.StatusCode >= http.StatusBadRequest {
		fields = append(fields, "error", fmt.Sprintf("http status: %d", response.StatusCode))
		t.slog.ErrorContext(request.Context(), message, fields...)
	} else {
		t.slog.InfoContext(request.Context(), message, fields...)
	}

	return nil
}

func (t *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	err := t.logRequest(request)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	response, err := t.rt.RoundTrip(request)

	err = t.logResponse(request, response, err, time.Since(start))
	return response, err
}

// CopyRequestBody read request body for logging
func CopyRequestBody(r *http.Request) ([]byte, error) {
	if r == nil || r.Body == nil {
		return []byte{}, nil
	}
	defer r.Body.Close()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(buf))

	return buf, nil
}

// CopyResponseBody read response body for logging
func CopyResponseBody(r *http.Response) ([]byte, error) {
	if r == nil || r.Body == nil {
		return []byte{}, nil
	}
	defer r.Body.Close()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(buf))

	return buf, nil
}
