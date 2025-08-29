// =============================================================================
// internal/helpers/utils/time_utils.go
// =============================================================================
package utils

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// TimePtr returns a pointer to the given time
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// TimeValue returns the time value from pointer, zero time if nil
func TimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// NowPtr returns a pointer to current time
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// IsExpired checks if a time is in the past
func IsExpired(t time.Time) bool {
	return time.Now().After(t)
}

// DurationBetween calculates duration between two times
func DurationBetween(start, end time.Time) time.Duration {
	if start.After(end) {
		return end.Sub(start)
	}
	return end.Sub(start)
}
