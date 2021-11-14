package addresses

import (
	"net"
	"net/http"
)

func GetRealIP(r *http.Request) *string {
	var remoteIP string

	if len(r.Header.Get("X-REAL-IP")) > 0 {
		remoteIP = r.Header.Get("X-REAL-IP")
	} else if len(r.Header.Get("X-Forwarded-For")) > 0 {
		remoteIP = r.Header.Get("X-Forwarded-For")
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remoteIP = r.RemoteAddr
		}
		remoteIP = host
	}
	return &remoteIP
}
