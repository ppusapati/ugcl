// =============================================================================
// internal/helpers/utils/slice_utils.go
// =============================================================================
package utils

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Remove removes a string from slice
func Remove(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// Unique returns unique strings from slice
func Unique(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, s := range slice {
		if !keys[s] {
			keys[s] = true
			result = append(result, s)
		}
	}
	return result
}

// Filter filters slice based on predicate function
func Filter(slice []string, predicate func(string) bool) []string {
	var result []string
	for _, s := range slice {
		if predicate(s) {
			result = append(result, s)
		}
	}
	return result
}

// Map transforms slice using mapper function
func Map(slice []string, mapper func(string) string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = mapper(s)
	}
	return result
}
