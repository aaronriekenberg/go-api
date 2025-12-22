package request

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/aaronriekenberg/go-api/config"
)

type IsExternal func(
	r *http.Request,
) (external bool)

func newExternalCheck(
	externalHost string,
) IsExternal {

	externalHost = strings.ToLower(externalHost)
	dotExternalHost := "." + externalHost

	return func(
		r *http.Request,
	) (external bool) {

		requestHost := strings.ToLower(r.Host)

		external = (requestHost == externalHost) || (strings.HasSuffix(requestHost, dotExternalHost))
		return
	}
}

var ExternalCheckInstance = sync.OnceValue(func() IsExternal {
	externalHost := config.Instance().RequestConfiguration.ExternalHost

	slog.Info("calling newExternalCheck",
		"externalHost", externalHost,
	)

	return newExternalCheck(externalHost)
})
