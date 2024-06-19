package versioninfo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
	"github.com/aaronriekenberg/go-api/version"
)

func versionInfoHandlerFunc() http.HandlerFunc {
	jsonBytes, err := json.Marshal(version.BuildInfoMap())
	if err != nil {
		panic(fmt.Errorf("versionInfoHandlerFunc json.Marshal error: %w", err))
	}

	return utils.JSONBytesHandlerFunc(jsonBytes)
}

func NewVersionInfoHandler() http.Handler {
	return versionInfoHandlerFunc()
}
