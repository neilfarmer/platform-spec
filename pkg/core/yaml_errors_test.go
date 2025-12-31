package core

import (
	"errors"
	"strings"
	"testing"
)

func TestEnhanceYAMLError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		specPath       string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:     "string instead of array",
			err:      errors.New("cannot unmarshal !!str `nginx` into []string"),
			specPath: "test.yaml",
			wantContains: []string{
				"YAML parsing error",
				"Expected a list",
				"got a single string value",
				"Wrong: packages: nginx",
				"Right: packages: [nginx]",
				"examples/basic.yaml",
			},
		},
		{
			name:     "array instead of string",
			err:      errors.New("cannot unmarshal !!seq into string"),
			specPath: "test.yaml",
			wantContains: []string{
				"Expected a string",
				"got a list",
				"Wrong: name: [test]",
				"Right: name: test",
			},
		},
		{
			name:     "string instead of int",
			err:      errors.New("cannot unmarshal !!str `8080` into int"),
			specPath: "test.yaml",
			wantContains: []string{
				"Expected a number",
				"got a string",
				"Wrong: port: \"8080\"",
				"Right: port: 8080",
			},
		},
		{
			name:     "string instead of bool",
			err:      errors.New("cannot unmarshal !!str `true` into bool"),
			specPath: "test.yaml",
			wantContains: []string{
				"Expected a boolean",
				"got a string",
				"Wrong: enabled: \"true\"",
				"Right: enabled: true",
			},
		},
		{
			name:     "map instead of array",
			err:      errors.New("cannot unmarshal !!map into []core.PackageTest"),
			specPath: "test.yaml",
			wantContains: []string{
				"Expected a list",
				"got a mapping/object",
				"this field should be a list",
			},
		},
		{
			name:     "unknown field with suggestion",
			err:      errors.New("yaml: unmarshal errors:\n  line 5: field packges not found in type core.PackageTest"),
			specPath: "test.yaml",
			wantContains: []string{
				"Unknown field 'packges'",
				"Did you mean 'packages'",
				"examples/basic.yaml",
			},
		},
		{
			name:     "line number extraction",
			err:      errors.New("yaml: line 23: cannot unmarshal !!str into []string"),
			specPath: "spec.yaml",
			wantContains: []string{
				"spec.yaml at line 23",
				"Expected a list",
			},
		},
		{
			name:     "generic fallback",
			err:      errors.New("some other yaml error"),
			specPath: "test.yaml",
			wantContains: []string{
				"YAML parsing error",
				"Troubleshooting tips",
				"Check indentation",
				"examples/basic.yaml",
				"platform-spec test --help",
			},
		},
		{
			name:     "nil error returns nil",
			err:      nil,
			specPath: "test.yaml",
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enhanceYAMLError(tt.err, tt.specPath)

			// If input error is nil, output should be nil
			if tt.err == nil {
				if got != nil {
					t.Errorf("enhanceYAMLError() with nil error should return nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("enhanceYAMLError() returned nil, want non-nil error")
				return
			}

			gotStr := got.Error()

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(gotStr, want) {
					t.Errorf("enhanceYAMLError() error should contain %q, got:\n%s", want, gotStr)
				}
			}

			// Check that unwanted strings are not present
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(gotStr, notWant) {
					t.Errorf("enhanceYAMLError() error should not contain %q, got:\n%s", notWant, gotStr)
				}
			}
		})
	}
}

func TestExtractLineNumber(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		want    string
	}{
		{
			name:   "standard yaml.v3 format",
			errMsg: "yaml: line 23: cannot unmarshal",
			want:   "23",
		},
		{
			name:   "alternate format",
			errMsg: "line 42: some error",
			want:   "42",
		},
		{
			name:   "no line number",
			errMsg: "some error without line number",
			want:   "",
		},
		{
			name:   "multiple line numbers - first one",
			errMsg: "line 5: error on line 10",
			want:   "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLineNumber(tt.errMsg)
			if got != tt.want {
				t.Errorf("extractLineNumber() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildLinePrefix(t *testing.T) {
	tests := []struct {
		name     string
		specPath string
		lineNum  string
		want     string
	}{
		{
			name:     "with line number",
			specPath: "test.yaml",
			lineNum:  "23",
			want:     "YAML parsing error in test.yaml at line 23:\n  ",
		},
		{
			name:     "without line number",
			specPath: "spec.yaml",
			lineNum:  "",
			want:     "YAML parsing error in spec.yaml:\n  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLinePrefix(tt.specPath, tt.lineNum)
			if got != tt.want {
				t.Errorf("buildLinePrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractFieldName(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   string
	}{
		{
			name:   "field not found pattern",
			errMsg: "field packges not found in type core.PackageTest",
			want:   "packges",
		},
		{
			name:   "different field name",
			errMsg: "field stat not found",
			want:   "stat",
		},
		{
			name:   "no field name found",
			errMsg: "some other error",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFieldName(tt.errMsg)
			if got != tt.want {
				t.Errorf("extractFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuggestFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      string
	}{
		{
			name:      "packges -> packages",
			fieldName: "packges",
			want:      "packages",
		},
		{
			name:      "stat -> state",
			fieldName: "stat",
			want:      "state",
		},
		{
			name:      "servce -> service",
			fieldName: "servce",
			want:      "service",
		},
		{
			name:      "enbled -> enabled",
			fieldName: "enbled",
			want:      "enabled",
		},
		{
			name:      "cmd -> command",
			fieldName: "cmd",
			want:      "command",
		},
		{
			name:      "no suggestion for unknown",
			fieldName: "unknownfield",
			want:      "",
		},
		{
			name:      "case insensitive",
			fieldName: "PACKGES",
			want:      "packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestFieldName(tt.fieldName)
			if got != tt.want {
				t.Errorf("suggestFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorMessageFormat(t *testing.T) {
	// Integration test: verify enhanced messages are more helpful than originals
	tests := []struct {
		name         string
		originalErr  error
		improvesOver string
	}{
		{
			name:         "type mismatch is more helpful",
			originalErr:  errors.New("cannot unmarshal !!str `nginx` into []string"),
			improvesOver: "!!str",
		},
		{
			name:         "includes examples",
			originalErr:  errors.New("cannot unmarshal !!seq into string"),
			improvesOver: "!!seq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := enhanceYAMLError(tt.originalErr, "test.yaml")
			enhancedStr := enhanced.Error()

			// Enhanced error should NOT contain cryptic YAML type syntax in the main message
			// (it can appear in "Original error:" section)
			lines := strings.Split(enhancedStr, "Original error:")
			mainMessage := lines[0]

			if strings.Contains(mainMessage, tt.improvesOver) {
				// This is okay only if it's in a quoted context showing what the error was
				if !strings.Contains(enhancedStr, "Common fix") && !strings.Contains(enhancedStr, "Troubleshooting") {
					t.Errorf("Enhanced error main message should not contain cryptic syntax %q, got:\n%s",
						tt.improvesOver, mainMessage)
				}
			}

			// Verify it contains helpful guidance
			if !strings.Contains(enhancedStr, "Wrong:") && !strings.Contains(enhancedStr, "Troubleshooting") {
				t.Errorf("Enhanced error should contain examples or troubleshooting tips, got:\n%s", enhancedStr)
			}
		})
	}
}
