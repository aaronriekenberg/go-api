package versioninfo

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aaronriekenberg/go-api/utils"
)

var buildInfoMap map[string]string

func versionInfoHandlerFunc() http.HandlerFunc {
	jsonBytes, err := json.Marshal(buildInfoMap)
	if err != nil {
		slog.Error("versionInfoHandlerFunc json.Marshal error",
			"error", err)
		os.Exit(1)
	}

	return utils.JSONBytesHandlerFunc(jsonBytes)
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
