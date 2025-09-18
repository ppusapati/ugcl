package helper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func ParseSubject(subject string) (prefix string, id string, err error) {
	parts := strings.SplitN(subject, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", "", fmt.Errorf("invalid subject format, expected 'type:id'")
	}
	return parts[0], parts[1], nil
}

var allowedPrefixes = map[string]bool{
	"user":     true,
	"role":     true,
	"team":     true,
	"branch":   true,
	"location": true,
}

var validIDPattern = regexp.MustCompile(`^[a-zA-Z0-9:_/-]+$`) // allow team:eng-dev-1 etc.

func ParseSubjectSafe(subject string) (prefix, id string, err error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "", "", errors.New("subject cannot be empty")
	}

	parts := strings.SplitN(subject, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid subject format, expected 'type:id'")
	}

	prefix = strings.ToLower(strings.TrimSpace(parts[0]))
	id = strings.TrimSpace(parts[1])

	if !allowedPrefixes[prefix] {
		return "", "", fmt.Errorf("unauthorized subject type: %s", prefix)
	}

	if !validIDPattern.MatchString(id) {
		return "", "", fmt.Errorf("invalid subject ID: %s", id)
	}

	return prefix, id, nil
}
