package utils

import (
	"testing"
)

func TestStoreAndLoad(t *testing.T) {
	var gsm GenericSyncMap[int, string]

	gsm.Store(1, "one")
	gsm.Store(2, "two")

	value, ok := gsm.Load(1)
	if ok != true || value != "one" {
		t.Errorf("gsm.Load(1) = (%v, %v) want (\"one\", true)", ok, value)
	}

	value, ok = gsm.Load(2)
	if ok != true || value != "two" {
		t.Errorf("gsm.Load(1) = (%v, %v) want (\"two\", true)", ok, value)
	}

	value, ok = gsm.Load(3)
	if ok != false || value != "" {
		t.Errorf("gsm.Load(1) = (%v, %v) want (\"\", false)", ok, value)
	}
}
