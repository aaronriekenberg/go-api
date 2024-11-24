package versioninfo

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
	"github.com/aaronriekenberg/go-api/version"
)

func versionInfoHandlerFunc() http.HandlerFunc {
	return utils.JSONBytesHandlerFunc(utils.MustMarshalJSON(version.BuildInfoMap()))
}

func NewVersionInfoHandler() http.Handler {
	return versionInfoHandlerFunc()
}
