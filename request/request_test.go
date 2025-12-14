package request

import (
	"net/http"
	"testing"
)

func TestIsExternal(
	t *testing.T,
) {

	isExternal := newExternalCheck("aaronr.digital")

	tests := map[string]struct {
		wantValue bool
	}{
		"aaronr.digital":           {wantValue: true},
		"notaaronr.digital":        {wantValue: false},
		"www.aaronr.digital":       {wantValue: true},
		"aaronr.digital.com":       {wantValue: false},
		"www.aaronr.digital.com":   {wantValue: false},
		"stuff.www.aaronr.digital": {wantValue: true},
		".aaronr.digital":          {wantValue: true},
		"Aaronr.Digital":           {wantValue: true},
		"Notaaronr.Digital":        {wantValue: false},
		"WWW.AARONR.DIGITAL":       {wantValue: true},
		"Aaronr.Digital.Com":       {wantValue: false},
		"Www.Aaronr.Dgital.Com":    {wantValue: false},
		"Stuff.Www.Aaronr.Digital": {wantValue: true},
		".AARONR.DIGITAL":          {wantValue: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := http.Request{
				Host: name,
			}
			value := isExternal(&r)

			if value != tc.wantValue {
				t.Fatalf("test: %q got value: %v want value %v", name, value, tc.wantValue)
			}
		})
	}
}
