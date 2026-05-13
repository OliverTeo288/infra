package observability

import "context"

// AuditHook is the contract lower-level packages depend on. They don't import
// observability directly — they accept an AuditHook in their function
// signatures and the CLI layer plumbs in a concrete implementation.
//
// The first argument after ctx is always the action name; remaining attrs
// are slog-style alternating key/value pairs. Callers include only the
// fields they actually have (e.g. resource_arn for create-style calls,
// role_name for IAM, bucket for S3) — no more empty placeholders.
//
// A nil AuditHook is a valid value and behaves as a no-op, so it is safe
// to pass into AWS adapters from tests without setting up logging.
type AuditHook func(ctx context.Context, action string, attrs ...any)

// AuditFunc returns an AuditHook backed by the package-level slog logger.
// Use this in the CLI layer to wire up audit logging:
//
//	flows.CreateGitopsRole(ctx, observability.AuditFunc(), ...)
func AuditFunc() AuditHook {
	return Audit
}

// Run safely invokes h, treating a nil hook as a no-op.
func (h AuditHook) Run(ctx context.Context, action string, attrs ...any) {
	if h == nil {
		return
	}
	h(ctx, action, attrs...)
}
