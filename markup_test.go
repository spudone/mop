package mop

import (
	"reflect"
	"testing"
)

func TestExtractTagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<tag>", "tag"},
		{"</tag>", "tag"},
		{"<>", ""},
		{"<t>", "t"},
		{"<no tag>", "no tag"},
		{"<invalid>", "invalid"}, // This is valid for extractTagName, but IsTag will reject it if not in map.
		{"</>", "/"},
	}

	for _, test := range tests {
		result := extractTagName(test.input)
		if result != test.expected {
			t.Errorf("extractTagName(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestTokenize(t *testing.T) {
	profile := &Profile{}
	profile.Colors.Gain = "green"
	profile.Colors.Loss = "red"
	profile.Colors.Tag = "yellow"
	profile.Colors.Header = "white"
	profile.Colors.Time = "white"
	profile.Colors.Default = "white"
	profile.Colors.RowShading = "black"

	markup := NewMarkup(profile)

	tests := []struct {
		input    string
		expected []string
	}{
		{
			"<green>Hello</>",
			[]string{"<green>", "Hello", "</>"},
		},
		{
			"Normal text",
			[]string{"Normal text"},
		},
		{
			"<green>Hello, <red>world!</>",
			[]string{"<green>", "Hello, ", "<red>", "world!", "</>"},
		},
	}

	for _, test := range tests {
		result := markup.Tokenize(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Tokenize(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestIsTag(t *testing.T) {
	profile := &Profile{}
	profile.Colors.Gain = "green"
	profile.Colors.Loss = "red"
	profile.Colors.Tag = "yellow"
	profile.Colors.Header = "white"
	profile.Colors.Time = "white"
	profile.Colors.Default = "white"
	profile.Colors.RowShading = "black"

	markup := NewMarkup(profile)

	tests := []struct {
		input    string
		expected bool
	}{
		{"<green>", true},
		{"</green>", true},
		{"<invalid>", false},
		{"<right>", true},
		{"hello", false},
	}

	for _, test := range tests {
		result := markup.IsTag(test.input)
		if result != test.expected {
			t.Errorf("IsTag(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
