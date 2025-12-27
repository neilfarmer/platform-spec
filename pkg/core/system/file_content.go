package system

import (
	"context"
	"fmt"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeFileContentTest executes a file content test
func executeFileContentTest(ctx context.Context, provider core.Provider, test core.FileContentTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// First check if file exists and is readable
	_, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("test -f %s && test -r %s", test.Path, test.Path))
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking file %s: %v", test.Path, err)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("File %s does not exist or is not readable", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	// Check contains strings
	if len(test.Contains) > 0 {
		for _, searchStr := range test.Contains {
			// Use grep -F for fixed string matching (no regex interpretation)
			_, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("grep -F %s %s >/dev/null 2>&1", core.ShellQuote(searchStr), test.Path))
			if err != nil {
				result.Status = core.StatusError
				result.Message = fmt.Sprintf("Error searching for '%s' in %s: %v", searchStr, test.Path, err)
				result.Duration = time.Since(start)
				return result
			}

			if exitCode != 0 {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("File %s does not contain '%s'", test.Path, searchStr)
				result.Details["missing"] = searchStr
				result.Duration = time.Since(start)
				return result
			}

			result.Details[searchStr] = "found"
		}
	}

	// Check matches regex pattern
	if test.Matches != "" {
		// Use grep -E for extended regex
		_, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("grep -E %s %s >/dev/null 2>&1", core.ShellQuote(test.Matches), test.Path))
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error matching pattern in %s: %v", test.Path, err)
			result.Duration = time.Since(start)
			return result
		}

		if exitCode != 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("File %s does not match pattern '%s'", test.Path, test.Matches)
			result.Details["pattern"] = test.Matches
			result.Duration = time.Since(start)
			return result
		}

		result.Details["pattern"] = "matched"
	}

	if len(test.Contains) > 0 && test.Matches != "" {
		result.Message = fmt.Sprintf("File %s contains all %d strings and matches pattern", test.Path, len(test.Contains))
	} else if len(test.Contains) > 0 {
		result.Message = fmt.Sprintf("File %s contains all %d strings", test.Path, len(test.Contains))
	} else if test.Matches != "" {
		result.Message = fmt.Sprintf("File %s matches pattern", test.Path)
	}

	result.Duration = time.Since(start)
	return result
}
