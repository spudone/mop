package mop

import "testing"

func TestFloat2Str(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100.0, "100.000"},
		{1500000.0, "1.500M"},       // > 1e6
		{2000000000.0, "2.000B"},    // > 1e9
		{3000000000000.0, "3.000T"}, // > 1e12
		{0.0, "0.000"},
		{-500.0, "-500.000"},
		{999.0, "999.000"},
		{1000.0, "1000.000"},
		{100001.0, "100.001K"}, // > 1.0e5
	}

	for _, test := range tests {
		result := float2Str(test.input)
		if result != test.expected {
			t.Errorf("float2Str(%f) = %q; want %q", test.input, result, test.expected)
		}
	}
}
