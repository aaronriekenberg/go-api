package request

import (
	"net/http"
	"strings"
)

func IsExternal(
	r *http.Request,
) bool {
	const externalHost = "aaronr.digital"
	const dotExternalHost = "." + externalHost

	requestHost := strings.ToLower(r.Host)

	return (requestHost == externalHost) || (strings.HasSuffix(requestHost, dotExternalHost))
}
