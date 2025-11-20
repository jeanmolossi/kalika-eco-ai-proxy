package httpapi

import "strings"

// extractAPIKey parses the Authorization header and returns the API key.
// It supports the "Bearer <token>" format and also accepts a raw value.
func extractAPIKey(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	// Bearer token format
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	// Fallback: treat the header as the raw key.
	return authHeader
}
