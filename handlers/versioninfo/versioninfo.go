package versioninfo

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
	"github.com/aaronriekenberg/go-api/version"
)

func NewVersionInfoHandler() http.Handler {
	return utils.JSONBytesHandlerFunc(utils.MustMarshalJSON(version.BuildInfoMap()))
}
