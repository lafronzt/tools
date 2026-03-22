// Package addresses provides helpers for extracting client IP addresses from
// HTTP requests that may pass through proxies or load balancers.
package addresses

import (
	"net"
	"net/http"
	"strings"
)

// GetRealIP returns the originating client IP from r, checking headers in
// order of specificity: X-Real-IP → X-Forwarded-For → RemoteAddr.
//
// When X-Forwarded-For contains a comma-separated list of addresses (as added
// by each successive proxy), only the first (originating) address is returned.
//
// The returned pointer is never nil; it points to an empty string if no
// address could be determined.
//
// Security: X-Real-IP and X-Forwarded-For are client-controlled headers. Only
// trust the value returned by this function when your server sits behind a
// trusted proxy or load balancer that strips or overwrites these headers before
// forwarding requests. Accepting them from untrusted sources allows IP spoofing.
func GetRealIP(r *http.Request) *string {
	var ip string

	switch {
	case r.Header.Get("X-Real-IP") != "":
		ip = firstIP(r.Header.Get("X-Real-IP"))
	case r.Header.Get("X-Forwarded-For") != "":
		ip = firstIP(r.Header.Get("X-Forwarded-For"))
	default:
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		} else {
			ip = host
		}
	}

	return &ip
}

// firstIP returns the first IP from a comma-separated list, trimmed of
// whitespace. If the value contains no comma it is returned as-is.
func firstIP(s string) string {
	if idx := strings.IndexByte(s, ','); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}
