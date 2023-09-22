package versioninfo

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aaronriekenberg/go-api/utils"
)

var buildInfoMap map[string]string

func versionInfoHandlerFunc() http.HandlerFunc {
	jsonBuffer, err := json.Marshal(buildInfoMap)
	if err != nil {
		slog.Error("versionInfoHandlerFunc json.Marshal error",
			"error", err)
		os.Exit(1)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(utils.ContentTypeHeaderKey, utils.ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBuffer))
	})
}

func NewVersionInfoHandler() http.Handler {
	return versionInfoHandlerFunc()
}

func init() {
	buildInfoMap = make(map[string]string)

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		buildInfoMap["GoVersion"] = buildInfo.GoVersion
		for _, setting := range buildInfo.Settings {
			if strings.HasPrefix(setting.Key, "vcs") {
				buildInfoMap[setting.Key] = setting.Value
			}
		}
	}
}
