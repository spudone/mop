package mop

import (
	"testing"
)

func TestStringToNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$10.50", 10.50},
		{"0.03%", 0.03},
		{"1.2K", 1200.0},
		{"1.5M", 1500000.0},
		{"2.0B", 2000000000.0},
		{"3.0T", 3000000000000.0},
		{" 100 ", 100.0},
		{"100", 100.0},
	}

	for _, test := range tests {
		result := stringToNumber(test.input)
		// Float comparison needs epsilon ideally, but these should be exact powers of 10 or exact floats.
		// Let's use a small epsilon if needed, but for now simple comparison might fail on precision.
		// However, stringToNumber uses strconv.ParseFloat(newString, 64).
		// 1.2E+3 is exactly 1200.0.
		if result != test.expected {
			t.Errorf("stringToNumber(%q) = %f; want %f", test.input, result, test.expected)
		}
	}
}

func TestFilterApply(t *testing.T) {
	profile := &Profile{}
	profile.Tickers = []string{"AAPL", "GOOG"}
	// Create a filter instance
	filter := NewFilter(profile)

	// Create some stocks
	stocks := []Stock{
		{Ticker: "AAPL", LastTrade: "$150.00", Change: "+2.00", ChangePct: "1.35%", Direction: 1},
		{Ticker: "GOOG", LastTrade: "$2800.00", Change: "-10.00", ChangePct: "-0.35%", Direction: -1},
		{Ticker: "MSFT", LastTrade: "$300.00", Change: "0.00", ChangePct: "0.00%", Direction: 0},
	}

	// Set filter expression: last > 200
	profile.SetFilter("last > 200")

	filtered := filter.Apply(stocks)

	expectedTickers := []string{"GOOG", "MSFT"}

	if len(filtered) != len(expectedTickers) {
		t.Errorf("Apply() returned %d stocks; want %d", len(filtered), len(expectedTickers))
	} else {
		for i, stock := range filtered {
			if stock.Ticker != expectedTickers[i] {
				t.Errorf("Apply() stock[%d] = %s; want %s", i, stock.Ticker, expectedTickers[i])
			}
		}
	}

    // Set filter expression: change > 0
    profile.SetFilter("change > 0")
    filtered = filter.Apply(stocks)
    expectedTickers = []string{"AAPL"}

    if len(filtered) != len(expectedTickers) {
        t.Errorf("Apply(change>0) returned %d stocks; want %d", len(filtered), len(expectedTickers))
    } else {
         if filtered[0].Ticker != "AAPL" {
             t.Errorf("Apply(change>0) stock[0] = %s; want AAPL", filtered[0].Ticker)
         }
    }

}
