package planning

import (
	"sort"
	"strings"
)

func LStringContainsCheck(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func LStringsUniqueGet(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func LStringsSortedGet(values []string) []string {
	result := LStringsUniqueGet(values)
	sort.Strings(result)
	return result
}
