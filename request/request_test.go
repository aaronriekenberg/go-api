package request_test

import (
	"net/http"
	"testing"

	"github.com/aaronriekenberg/go-api/request"

	"github.com/google/go-cmp/cmp"
)

func TestIsExternal(
	t *testing.T,
) {
	tests := map[string]struct {
		host      string
		wantValue bool
	}{
		"aaronr.digital":           {host: "aaronr.digital", wantValue: true},
		"notaaronr.digital":        {host: "notaaronr.digital", wantValue: false},
		"www.aaronr.digital":       {host: "www.aaronr.digital", wantValue: true},
		"stuff.www.aaronr.digital": {host: "stuff.www.aaronr.digital", wantValue: true},
		".www.aaronr.digital":      {host: ".www.aaronr.digital", wantValue: true},
		"Aaronr.Digital":           {host: "Aaronr.Digital", wantValue: true},
		"NotAaronr.Digital":        {host: "NotAaronr.Digital", wantValue: false},
		"Www.Aaronr.Digital":       {host: "Www.Aaronr.Digital", wantValue: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := http.Request{
				Host: tc.host,
			}
			value := request.IsExternal(&r)

			diff := cmp.Diff(tc.wantValue, value)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
