package mop

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestIsSupportedColor(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"green", true},
		{"red", true},
		{"black", true},
		{"lightgray", true},
		{"purple", false}, // magenta is supported, purple is not
		{"orange", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsSupportedColor(test.input)
		if result != test.expected {
			t.Errorf("IsSupportedColor(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestAddTickers(t *testing.T) {
	// Use a temporary file for the profile.
	tmpFile, err := ioutil.TempFile("", "mop_test_profile_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	profile := &Profile{
		filename: tmpFile.Name(),
		Tickers:  []string{"AAPL", "GOOG"},
	}

	newTickers := []string{"MSFT", "AAPL"} // One new, one existing
	added, err := profile.AddTickers(newTickers)

	if err != nil {
		t.Errorf("AddTickers returned error: %v", err)
	}

	if added != 1 {
		t.Errorf("AddTickers added %d; want 1", added)
	}

	expectedTickers := []string{"AAPL", "GOOG", "MSFT"}
	sort.Strings(expectedTickers)
	// AddTickers sorts the tickers

	if !reflect.DeepEqual(profile.Tickers, expectedTickers) {
		t.Errorf("Tickers = %v; want %v", profile.Tickers, expectedTickers)
	}
}

func TestRemoveTickers(t *testing.T) {
	// Use a temporary file for the profile.
	tmpFile, err := ioutil.TempFile("", "mop_test_profile_remove_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	profile := &Profile{
		filename: tmpFile.Name(),
		Tickers:  []string{"AAPL", "GOOG", "MSFT"},
	}

	toRemove := []string{"GOOG", "TSLA"} // One existing, one non-existing
	removed, err := profile.RemoveTickers(toRemove)

	if err != nil {
		t.Errorf("RemoveTickers returned error: %v", err)
	}

	if removed != 1 {
		t.Errorf("RemoveTickers removed %d; want 1", removed)
	}

	expectedTickers := []string{"AAPL", "MSFT"}
	// RemoveTickers preserves order (it uses slice append)

	if !reflect.DeepEqual(profile.Tickers, expectedTickers) {
		t.Errorf("Tickers = %v; want %v", profile.Tickers, expectedTickers)
	}
}
