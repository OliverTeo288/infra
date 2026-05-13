package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Context-key types are unexported to avoid collisions; see
// https://pkg.go.dev/context#WithValue.
type (
	sessionIDKey struct{}
	attrsKey     struct{}
)

var (
	loggerOnce    sync.Once
	defaultLogger *slog.Logger
)

// Init configures the package-level slog logger.
//
//	format: "json" (production / shippable to Datadog) or "text" (TTY default)
//	level:  "debug" | "info" | "warn" | "error"
//
// Reads INFRA_LOG_FORMAT and INFRA_LOG_LEVEL env vars as overrides. Safe
// to call multiple times — replaces the package logger atomically.
func Init(format, level string) {
	if v := os.Getenv("INFRA_LOG_FORMAT"); v != "" {
		format = v
	}
	if v := os.Getenv("INFRA_LOG_LEVEL"); v != "" {
		level = v
	}

	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// L returns the package-level logger; safe to call before Init (returns a
// no-op-style discard logger on first use).
func L() *slog.Logger {
	loggerOnce.Do(func() {
		if defaultLogger == nil {
			defaultLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
		}
	})
	return defaultLogger
}

// WithSession attaches a per-invocation random session ID to the context.
// Subsequent calls to FromContext(ctx).Info(...) will include this ID
// so a single `infra init` flow's logs can be correlated end-to-end.
func WithSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, newSessionID())
}

// WithAttrs returns a derived context that carries the supplied slog-style
// key/value attributes. FromContext(ctx) will automatically include them
// on every log line until you derive a new context.
//
// Use this in flows to attach widely-relevant fields once (account_id,
// region, role_name) instead of repeating them at every Audit call site.
func WithAttrs(ctx context.Context, attrs ...any) context.Context {
	existing, _ := ctx.Value(attrsKey{}).([]any)
	// Defensive copy so callers can't mutate prior attrs via slice aliasing.
	combined := make([]any, 0, len(existing)+len(attrs))
	combined = append(combined, existing...)
	combined = append(combined, attrs...)
	return context.WithValue(ctx, attrsKey{}, combined)
}

// FromContext returns a logger pre-populated with any contextual attributes
// (session ID and anything attached via WithAttrs). Use this for any
// structured log line emitted from inside a command flow.
func FromContext(ctx context.Context) *slog.Logger {
	logger := L()
	if id, ok := ctx.Value(sessionIDKey{}).(string); ok {
		logger = logger.With("session_id", id)
	}
	if attrs, ok := ctx.Value(attrsKey{}).([]any); ok && len(attrs) > 0 {
		logger = logger.With(attrs...)
	}
	return logger
}

// Audit emits a single structured log line at INFO describing an AWS-mutating
// action. The first attribute is always "action"; callers append only the
// attributes they actually have. Context attributes (session_id, account_id,
// region, etc. attached via WithAttrs) are included automatically.
//
// Example:
//
//	observability.Audit(ctx, "iam:CreateRole",
//	    "resource_arn", aws.ToString(out.Role.Arn),
//	    "role_name", roleName,
//	)
func Audit(ctx context.Context, action string, attrs ...any) {
	all := append([]any{"action", action}, attrs...)
	FromContext(ctx).Info("aws_action", all...)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newSessionID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
