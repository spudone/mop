package mop

import "testing"

func TestCurrencyParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"$10.50", 10.50},
		{"10.50%", 10.50},
		{"100", 100.0},
		{"$0.00", 0.0},
	}

	for _, test := range tests {
		result := c(test.input)
		if result != test.expected {
			t.Errorf("c(%q) = %f; want %f", test.input, result, test.expected)
		}
	}
}

func TestMarketCapParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"100M", 100000000.0},
		{"1.5B", 1500000000.0},
		{"2T", 2000000000000.0},
		{"500K", 500000.0},
		{"100", 100.0},
		{"", 0.0},
	}

	for _, test := range tests {
		result := m(test.input)
		if result != test.expected {
			t.Errorf("m(%q) = %f; want %f", test.input, result, test.expected)
		}
	}
}
