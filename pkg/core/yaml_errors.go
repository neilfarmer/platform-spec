package core

import (
	"fmt"
	"regexp"
	"strings"
)

// enhanceYAMLError takes a yaml.v3 error and enhances it with helpful context,
// hints, and suggestions to make it more user-friendly.
func enhanceYAMLError(err error, specPath string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	lineNum := extractLineNumber(errMsg)
	linePrefix := buildLinePrefix(specPath, lineNum)

	// Pattern 1: String instead of array (e.g., "cannot unmarshal !!str into []string")
	if strings.Contains(errMsg, "cannot unmarshal !!str") && strings.Contains(errMsg, "into []") {
		return fmt.Errorf("%sProblem: Expected a list, but got a single string value\n\n"+
			"Common fix: Wrap single values in brackets\n"+
			"  Wrong: packages: nginx\n"+
			"  Right: packages: [nginx]\n\n"+
			"See examples/basic.yaml for reference\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Pattern 2: Array instead of string (e.g., "cannot unmarshal !!seq into string")
	if strings.Contains(errMsg, "cannot unmarshal !!seq into string") {
		return fmt.Errorf("%sProblem: Expected a string, but got a list\n\n"+
			"Common fix: Remove brackets for single values\n"+
			"  Wrong: name: [test]\n"+
			"  Right: name: test\n\n"+
			"See examples/basic.yaml for reference\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Pattern 3: Integer as string (e.g., "cannot unmarshal !!str into int")
	if strings.Contains(errMsg, "cannot unmarshal !!str") && strings.Contains(errMsg, "into int") {
		return fmt.Errorf("%sProblem: Expected a number, but got a string\n\n"+
			"Common fix: Remove quotes around numbers\n"+
			"  Wrong: port: \"8080\"\n"+
			"  Right: port: 8080\n\n"+
			"See examples/basic.yaml for reference\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Pattern 4: Boolean as string (e.g., "cannot unmarshal !!str into bool")
	if strings.Contains(errMsg, "cannot unmarshal !!str") && strings.Contains(errMsg, "into bool") {
		return fmt.Errorf("%sProblem: Expected a boolean (true/false), but got a string\n\n"+
			"Common fix: Remove quotes around boolean values\n"+
			"  Wrong: enabled: \"true\"\n"+
			"  Right: enabled: true\n\n"+
			"See examples/basic.yaml for reference\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Pattern 5: Map/object instead of array (e.g., "cannot unmarshal !!map into []")
	if strings.Contains(errMsg, "cannot unmarshal !!map") && strings.Contains(errMsg, "into []") {
		return fmt.Errorf("%sProblem: Expected a list, but got a mapping/object\n\n"+
			"Common fix: Check your YAML structure - this field should be a list\n"+
			"  Wrong:\n"+
			"    tests:\n"+
			"      packages: value\n"+
			"  Right:\n"+
			"    tests:\n"+
			"      packages:\n"+
			"        - name: test\n\n"+
			"See examples/basic.yaml for reference\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Pattern 6: Unknown/invalid field (e.g., "field xyz not found in type")
	if strings.Contains(errMsg, "field") && (strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "cannot unmarshal")) {
		fieldName := extractFieldName(errMsg)
		suggestion := suggestFieldName(fieldName)
		if suggestion != "" {
			return fmt.Errorf("%sProblem: Unknown field '%s'\n\n"+
				"Did you mean '%s'?\n\n"+
				"See examples/basic.yaml for all valid fields\n\n"+
				"Original error: %v", linePrefix, fieldName, suggestion, err)
		}
		return fmt.Errorf("%sProblem: Unknown or invalid field in YAML structure\n\n"+
			"See examples/basic.yaml for all valid fields\n\n"+
			"Original error: %v", linePrefix, err)
	}

	// Generic fallback - add helpful troubleshooting tips
	return fmt.Errorf("%s%v\n\n"+
		"Troubleshooting tips:\n"+
		"  - Check indentation (YAML uses 2 spaces, not tabs)\n"+
		"  - Verify quotes around strings with special characters\n"+
		"  - Ensure colons have a space after them (key: value)\n"+
		"  - See examples/basic.yaml for correct format\n"+
		"  - Run: platform-spec test --help", linePrefix, err)
}

// extractLineNumber extracts the line number from yaml.v3 error messages
// which typically have format "line X: ..." or "yaml: line X: ..."
func extractLineNumber(errMsg string) string {
	lineRegex := regexp.MustCompile(`line (\d+)`)
	matches := lineRegex.FindStringSubmatch(errMsg)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// buildLinePrefix creates a consistent prefix for error messages
// that includes the spec file path and line number if available
func buildLinePrefix(specPath, lineNum string) string {
	if lineNum != "" {
		return fmt.Sprintf("YAML parsing error in %s at line %s:\n  ", specPath, lineNum)
	}
	return fmt.Sprintf("YAML parsing error in %s:\n  ", specPath)
}

// extractFieldName tries to extract field name from error messages
func extractFieldName(errMsg string) string {
	// Try pattern: "field xyz not found"
	fieldRegex := regexp.MustCompile(`field (\w+) not found`)
	matches := fieldRegex.FindStringSubmatch(errMsg)
	if len(matches) >= 2 {
		return matches[1]
	}

	// Try pattern: yaml key extraction from various formats
	// This is a best-effort attempt
	parts := strings.Split(errMsg, ":")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 && len(part) < 30 && !strings.Contains(part, " ") {
			return part
		}
	}

	return ""
}

// suggestFieldName provides suggestions for common field name typos
func suggestFieldName(fieldName string) string {
	// Common typos mapping
	suggestions := map[string]string{
		"packges":    "packages",
		"package":    "packages",
		"stat":       "state",
		"stats":      "state",
		"servce":     "service",
		"servcie":    "service",
		"svc":        "service",
		"contaner":   "container",
		"containr":   "container",
		"enbled":     "enabled",
		"enbaled":    "enabled",
		"flepath":    "path",
		"filepath":   "path",
		"usr":        "user",
		"usre":       "user",
		"grp":        "group",
		"group":      "group",
		"verion":     "version",
		"versoin":    "version",
		"comand":     "command",
		"commnad":    "command",
		"cmd":        "command",
		"restrt":     "restart",
		"restat":     "restart",
		"namspace":   "namespace",
		"namesapce":  "namespace",
		"deployemnt": "deployment",
		"deplomyent": "deployment",
		"configmap":  "configmap",
		"conifgmap":  "configmap",
	}

	// Convert to lowercase for matching
	lower := strings.ToLower(fieldName)
	if suggestion, ok := suggestions[lower]; ok {
		return suggestion
	}

	return ""
}
