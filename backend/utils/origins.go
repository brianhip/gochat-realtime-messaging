package utils

import "strings"

// DefaultAllowedOrigin is used when ALLOWED_ORIGINS is not set, matching
// the origin of the frontend dev server (npm start).
const DefaultAllowedOrigin = "http://localhost:3000"

// ParseAllowedOrigins builds an exact-match set of allowed origins from a
// comma-separated env var value, falling back to DefaultAllowedOrigin when
// the value is empty. The set is shared by the CORS middleware and the
// WebSocket origin check so both trust the same frontends.
func ParseAllowedOrigins(raw string) map[string]bool {
	if strings.TrimSpace(raw) == "" {
		raw = DefaultAllowedOrigin
	}

	origins := make(map[string]bool)
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = true
		}
	}

	return origins
}
