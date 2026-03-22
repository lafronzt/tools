package addresses

import (
	"net/http"
	"testing"
)

func newRequest(xRealIP, xForwardedFor, remoteAddr string) *http.Request {
	r := &http.Request{
		Header:     make(http.Header),
		RemoteAddr: remoteAddr,
	}
	if xRealIP != "" {
		r.Header.Set("X-Real-IP", xRealIP)
	}
	if xForwardedFor != "" {
		r.Header.Set("X-Forwarded-For", xForwardedFor)
	}
	return r
}

func TestGetRealIP(t *testing.T) {
	tests := []struct {
		name          string
		xRealIP       string
		xForwardedFor string
		remoteAddr    string
		want          string
	}{
		{
			name:       "RemoteAddr with port",
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Real-IP takes precedence over RemoteAddr",
			xRealIP:    "10.0.0.1",
			remoteAddr: "192.168.1.1:9999",
			want:       "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For single IP",
			xForwardedFor: "203.0.113.5",
			remoteAddr:    "10.0.0.1:80",
			want:          "203.0.113.5",
		},
		{
			name:          "X-Forwarded-For multiple IPs returns first",
			xForwardedFor: "203.0.113.5, 70.41.3.18, 150.172.238.178",
			remoteAddr:    "10.0.0.1:80",
			want:          "203.0.113.5",
		},
		{
			name:          "X-Real-IP takes precedence over X-Forwarded-For",
			xRealIP:       "10.0.0.1",
			xForwardedFor: "203.0.113.5",
			remoteAddr:    "192.168.1.1:80",
			want:          "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For with whitespace",
			xForwardedFor: "  203.0.113.5  ,  70.41.3.18  ",
			remoteAddr:    "10.0.0.1:80",
			want:          "203.0.113.5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRequest(tc.xRealIP, tc.xForwardedFor, tc.remoteAddr)
			got := GetRealIP(r)
			if got == nil {
				t.Fatal("GetRealIP returned nil pointer")
			}
			if *got != tc.want {
				t.Errorf("GetRealIP() = %q, want %q", *got, tc.want)
			}
		})
	}
}
