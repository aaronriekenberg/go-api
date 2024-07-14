package staticfile

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/utils"
)

func NewStaticFileHandler(
	staticFileConfiguraton config.StaticFileConfiguration,
) http.Handler {

	fileServer := http.FileServer(
		dotFileHidingFileSystem{
			FileSystem: http.Dir(staticFileConfiguraton.RootPath),
		},
	)

	// TODO: make regexes configurable
	vnstatPNGRegex := regexp.MustCompile(`^/?vnstat/.*\.png$`)

	aaronrHostRegex := regexp.MustCompile(`^aaronr.digital|.*\.aaronr.digital$`)

	cacheControlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.Default().With(
			"urlPath", r.URL.Path,
		)

		logger.Debug("in cacheControlHandler")

		switch {
		case vnstatPNGRegex.MatchString(r.URL.Path):
			logger.Debug("vnstatPNGRegex matches")
			w.Header().Set(utils.CacheControlHeaderKey, "public, max-age=150")

		case aaronrHostRegex.MatchString(r.Host):
			logger.Debug("aaronrHostRegex matches")
			w.Header().Set(utils.CacheControlHeaderKey, "public, max-age=86400")

		default:
			logger.Debug("default case")
			w.Header().Set(utils.CacheControlHeaderKey, utils.CacheControlNoCache)
		}

		fileServer.ServeHTTP(w, r)
	})

	return cacheControlHandler
}
