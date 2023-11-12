package versioninfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/aaronriekenberg/go-api/utils"
)

func buildInfoMap() map[string]string {
	buildInfoMap := make(map[string]string)

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		buildInfoMap["GoVersion"] = buildInfo.GoVersion
		for _, setting := range buildInfo.Settings {
			if strings.HasPrefix(setting.Key, "GO") ||
				strings.HasPrefix(setting.Key, "vcs") {
				buildInfoMap[setting.Key] = setting.Value
			}
		}
	}

	return buildInfoMap
}

func versionInfoHandlerFunc() (http.HandlerFunc, error) {
	jsonBytes, err := json.Marshal(buildInfoMap())
	if err != nil {
		return nil, fmt.Errorf("versionInfoHandlerFunc json.Marshal error: %w", err)
	}

	return utils.JSONBytesHandlerFunc(jsonBytes), nil
}

func NewVersionInfoHandler() (http.Handler, error) {
	return versionInfoHandlerFunc()
}
