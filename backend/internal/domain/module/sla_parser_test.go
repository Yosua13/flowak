package module_test

import (
	"testing"

	"backend/internal/domain/module"
)

func TestParseSLA(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue int
		wantUnit  string
	}{
		{name: "hours indonesian", input: "24 jam", wantValue: 24, wantUnit: "jam"},
		{name: "days english", input: "3 days", wantValue: 3, wantUnit: "hari"},
		{name: "minutes english", input: "15 minutes", wantValue: 15, wantUnit: "menit"},
		{name: "weeks indonesian", input: "2 minggu", wantValue: 2, wantUnit: "minggu"},
		{name: "invalid text", input: "instant", wantValue: 0, wantUnit: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, unit := module.ParseSLA(tt.input)
			if val != tt.wantValue || unit != tt.wantUnit {
				t.Fatalf("ParseSLA(%q) = (%d, %q), want (%d, %q)", tt.input, val, unit, tt.wantValue, tt.wantUnit)
			}
		})
	}
}

func TestParseSLAFloatAndAny(t *testing.T) {
	fVal, fUnit := module.ParseSLAFloat("1,5 jam")
	if fVal != 1.5 || fUnit != "jam" {
		t.Fatalf("ParseSLAFloat(\"1,5 jam\") = (%v, %v), want (1.5, \"jam\")", fVal, fUnit)
	}

	anyVal, anyUnit := module.ParseSLAAny("2 days")
	if anyVal != 2.0 || anyUnit != "hari" {
		t.Fatalf("ParseSLAAny(\"2 days\") = (%v, %v), want (2.0, \"hari\")", anyVal, anyUnit)
	}

	nilVal, nilUnit := module.ParseSLAAny("invalid")
	if nilVal != nil || nilUnit != nil {
		t.Fatalf("ParseSLAAny(\"invalid\") = (%v, %v), want (nil, nil)", nilVal, nilUnit)
	}
}
