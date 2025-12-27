package kubernetes

import "strings"

// contains checks if a string contains a substring (test helper)
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
