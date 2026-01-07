package versioning

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

type APIVersion struct {
	Major int
	Minor int
}

// ParseVersion "v1.2" -> APIVersion{1, 2}
func ParseVersion(header string) APIVersion {
	if header == "" {
		return APIVersion{Major: 1, Minor: 0}
	}

	clean := strings.TrimPrefix(header, "v")
	parts := strings.Split(clean, ".")

	major := 1
	minor := 0

	if len(parts) > 0 {
		if v, err := strconv.Atoi(parts[0]); err == nil {
			major = v
		}
	}
	if len(parts) > 1 {
		if v, err := strconv.Atoi(parts[1]); err == nil {
			minor = v
		}
	}

	return APIVersion{Major: major, Minor: minor}
}

type contextKey string

const keyAPIVersion contextKey = "api_version"

// VersionMiddleware injects the API version into the request context
func VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := ParseVersion(r.Header.Get("X-API-Version"))
		ctx := context.WithValue(r.Context(), keyAPIVersion, version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
