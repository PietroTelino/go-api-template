package audit

import (
	"net"
	"net/http"
)

func ClientInfoFromRequest(request *http.Request) (ipAddress *string, userAgent *string) {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil && host != "" {
		ipAddress = &host
	}

	if requestUserAgent := request.UserAgent(); requestUserAgent != "" {
		userAgent = &requestUserAgent
	}

	return ipAddress, userAgent
}
