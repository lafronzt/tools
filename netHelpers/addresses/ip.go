// Address is a helper package for handling IP Addresses
package addresses

import (
	"net"
	"net/http"
	"strings"
)

// GetIP returns the IP address of the client
func GetRealIP(r *http.Request) *string {
	var headerValue string
	var remoteIP string

	if len(r.Header.Get("X-REAL-IP")) > 0 {
		headerValue = r.Header.Get("X-REAL-IP")
	} else if len(r.Header.Get("X-Forwarded-For")) > 0 {
		headerValue = r.Header.Get("X-Forwarded-For")
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remoteIP = r.RemoteAddr
		}
		remoteIP = host
	}

	if len(headerValue) > 0 {
		remoteIP = splitStringAndTakeFirstValue(headerValue)
	}

	return &remoteIP
}

func splitStringAndTakeFirstValue(str string) string {
    if strings.Contains(str, ",") {
        return strings.Split(str, ",")[0]
    }
    return str
}
