package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
)

// AuthProvider fetches an Authorization header value (e.g. "Bearer <token>") for a request.
// It may consult context, request, refresh tokens, etc. If it returns "", no Authorization header is set.
// If it returns an error the RoundTrip will return that error.
type AuthProvider func(req *http.Request) (authHeaderValue string, err error)

// Options controls the behavior of the CustomRoundTripper.
type Options struct {
	// Underlying transport to call. If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// Static headers to add to every request (GitHub-style headers or others).
	// Values will be set on req.Header (overwrites any existing header with same name).
	StaticHeaders map[string]string

	// Optional function called to provide Authorization header per-request.
	AuthProvider AuthProvider

	// Logger used for structured logging. If nil, slog.Default() is used.
	Logger *slog.Logger

	// Maximum number of bytes to log for request and response bodies.
	// Set to 0 to disable body logging.
	MaxBodyLogBytes int64
}

// CustomRoundTripper implements http.RoundTripper
type CustomRoundTripper struct {
	base            http.RoundTripper
	staticHeaders   map[string]string
	authProvider    AuthProvider
	logger          *slog.Logger
	maxBodyLogBytes int64
}

// NewCustomRoundTripper constructs a CustomRoundTripper with sane defaults.
func NewCustomRoundTripper(opts Options) *CustomRoundTripper {
	base := opts.Base
	if base == nil {
		base = http.DefaultTransport
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// copy static headers to avoid mutation later
	static := map[string]string{}
	for k, v := range opts.StaticHeaders {
		static[k] = v
	}

	return &CustomRoundTripper{
		base:            base,
		staticHeaders:   static,
		authProvider:    opts.AuthProvider,
		logger:          logger,
		maxBodyLogBytes: opts.MaxBodyLogBytes,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (c *CustomRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Create a shallow clone of request to avoid mutating caller's request headers/body
	req2 := req.Clone(req.Context())

	// Inject static headers (e.g., GitHub headers)
	for k, v := range c.staticHeaders {
		req2.Header.Set(k, v)
	}

	// Inject auth header if provider present
	if c.authProvider != nil {
		val, err := c.authProvider(req2)
		if err != nil {
			c.logger.Error("auth provider error", slog.String("method", req2.Method), slog.String("url", req2.URL.String()), slog.Any("error", err))
			return nil, err
		}
		if val != "" {
			req2.Header.Set("Authorization", val)
		}
	}

	c.logger.Info("HTTP Request",
		slog.String("method", req2.Method),
		slog.String("url", req2.URL.String()),
	)

	// Perform the actual request
	resp, err := c.base.RoundTrip(req2)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("HTTP Error",
			slog.String("method", req2.Method),
			slog.String("url", req2.URL.String()),
			slog.Any("error", err),
			slog.Duration("took", duration),
		)
		return nil, err
	}

	c.logger.Info("HTTP Response",
		slog.Int("status", resp.StatusCode),
		slog.String("method", req2.Method),
		slog.String("url", req2.URL.String()),
		slog.Duration("took", duration),
	)

	return resp, nil
}

// Helper for simple API: create a transport that injects GitHub headers and a token from TokenManager
// Accepts a context with TokenManager, and logger.
func NewGithubStyleTransport(ctx context.Context, logger *slog.Logger) *CustomRoundTripper {
	static := map[string]string{
		"Accept":               "application/vnd.github+json",
		"X-GitHub-Api-Version": "2022-11-28",
	}

	authProv := func(req *http.Request) (string, error) {
		// Try to get TokenManager from context first
		if tm := ctx.Value(config.TokenManagerKey); tm != nil {
			if tokenManager, ok := tm.(interface{ GetCurrentToken() string }); ok {
				token := tokenManager.GetCurrentToken()
				if token == "" {
					return "", nil
				}
				return "Bearer " + token, nil
			}
		}

		// Fallback: if no TokenManager, try to get tokens array directly (use first token)
		if tokensValue := ctx.Value(config.TokenKey); tokensValue != nil {
			if tokens, ok := tokensValue.([]string); ok && len(tokens) > 0 && tokens[0] != "" {
				return "Bearer " + tokens[0], nil
			}
		}

		return "", nil
	}

	return NewCustomRoundTripper(Options{
		Base:            http.DefaultTransport,
		StaticHeaders:   static,
		AuthProvider:    authProv,
		Logger:          logger,
		MaxBodyLogBytes: 0,
	})
}

// ErrNoAuth is a sample sentinel error
var ErrNoAuth = errors.New("no auth available")
