package middleware

import "context"

// ID is a short alias for RequestIDFromContext (defined in middleware.go),
// kept because it's the name already used across the handler packages.
func ID(ctx context.Context) string { return RequestIDFromContext(ctx) }
