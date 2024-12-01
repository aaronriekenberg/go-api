package request

import (
	"net/http"
	"regexp"
)

var externalHostRegex = regexp.MustCompile(`^aaronr.digital|.*\.aaronr.digital$`)

func RequestIsExternal(
	r *http.Request,
) bool {
	return externalHostRegex.MatchString(r.Host)
}
