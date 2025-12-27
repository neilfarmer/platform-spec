package core_test

import (
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "string with single quote",
			input:    "it's",
			expected: "'it'\\''s'",
		},
		{
			name:     "string with multiple single quotes",
			input:    "can't won't don't",
			expected: "'can'\\''t won'\\''t don'\\''t'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "special characters",
			input:    "hello$world",
			expected: "'hello$world'",
		},
		{
			name:     "backslash",
			input:    "hello\\world",
			expected: "'hello\\world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.ShellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("ShellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
