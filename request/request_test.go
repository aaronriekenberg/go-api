package request_test

import (
	"net/http"
	"testing"

	"github.com/aaronriekenberg/go-api/request"
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
		"aaronr.digital.com":       {host: "aaronr.digital.com", wantValue: false},
		"www.aaronr.digital.com":   {host: "www.aaronr.digital.com", wantValue: false},
		"stuff.www.aaronr.digital": {host: "stuff.www.aaronr.digital", wantValue: true},
		".aaronr.digital":          {host: ".aaronr.digital", wantValue: true},
		"Aaronr.Digital":           {host: "Aaronr.Digital", wantValue: true},
		"Aaronr.Digital.Com":       {host: "Aaronr.Digital.Com", wantValue: false},
		"Www.Aaronr.Digital.Com":   {host: "Www.Aaronr.Digital.Com", wantValue: false},
		"NotAaronr.Digital":        {host: "NotAaronr.Digital", wantValue: false},
		"Www.Aaronr.Digital":       {host: "Www.Aaronr.Digital", wantValue: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if name != tc.host {
				t.Fatalf("test name: %q host: %q name != tc.host", name, tc.host)

			}

			r := http.Request{
				Host: tc.host,
			}
			value := request.IsExternal(&r)

			if value != tc.wantValue {
				t.Fatalf("test name: %q host: %q got value: %v want value %v", name, tc.host, value, tc.wantValue)
			}
		})
	}
}
