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

	dotExternalHost := "." + externalHost

	return func(
		r *http.Request,
	) (external bool) {

		requestHost := strings.ToLower(r.Host)

		external = (requestHost == externalHost) || (strings.HasSuffix(requestHost, dotExternalHost))
		return
	}
}

var ExternalCheck = sync.OnceValue(func() IsExternal {
	externalHost := config.ConfigurationInstance().RequestConfiguration.ExternalHost

	slog.Info("Calling newExternalCheck",
		"externalHost", externalHost,
	)

	return newExternalCheck(externalHost)
})
