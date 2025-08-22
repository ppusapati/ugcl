package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"p9e.in/ugcl/packages/models"
)

// Security context structure
type SecurityContext struct {
	Timestamp time.Time
	Username  string
}

// Validation errors
var (
	ErrInvalidQuery      = errors.New("invalid query")
	ErrUnsafeCharacters  = errors.New("query contains unsafe characters")
	ErrEmptyIdentifier   = errors.New("empty identifier")
	ErrInvalidIdentifier = errors.New("invalid identifier character")
	ErrEmptyTableName    = errors.New("empty table name")
	ErrInvalidTableName  = errors.New("invalid characters in table name")
	ErrEmptyFieldName    = errors.New("empty field name")
	ErrInvalidFieldName  = errors.New("invalid characters in field name")
	ErrMaxLengthExceeded = errors.New("maximum length exceeded")
)

// Validation constants
const (
	MaxQueryLength      = 4096
	MaxIdentifierLength = 63 // PostgreSQL's maximum identifier length
	MaxWhereLength      = 1000
)

// Regular expressions for validation
var (
	validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	sqlKeywordsRegex     = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|TRUNCATE|ALTER|UNION|EXEC|EXECUTE)\b`)
	unsafePatterns       = []string{
		"--",                 // SQL comment
		"/*",                 // Multi-line comment start
		"*/",                 // Multi-line comment end
		"xp_",                // Extended stored procedures
		"exec",               // Execute statement
		"execute",            // Execute statement
		"sp_",                // Stored procedure
		"sysobjects",         // System objects
		"syscolumns",         // System columns
		"waitfor",            // Time delay
		"benchmark",          // Performance testing
		"sleep",              // Time delay
		"information_schema", // System schema
		"@@",                 // System variables
	}
)

// NewSecurityContext creates a new security context with current timestamp and username
func NewSecurityContext(username string) *SecurityContext {
	return &SecurityContext{
		Timestamp: time.Now().UTC(),
		Username:  username,
	}
}

// ValidateQuery checks for common SQL injection patterns with security context
func ValidateQuery(query string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating query: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		query)

	// Check query length
	if len(query) > MaxQueryLength {
		return fmt.Errorf("%w: query length %d exceeds maximum %d",
			ErrMaxLengthExceeded, len(query), MaxQueryLength)
	}

	query = strings.ToLower(query)

	// Check for multiple statements
	if strings.Contains(query, ";") {
		return fmt.Errorf("%w: multiple statements detected", ErrInvalidQuery)
	}

	// Check for SQL keywords in unexpected places
	if sqlKeywordsRegex.MatchString(query) {
		return fmt.Errorf("%w: unauthorized SQL keywords detected", ErrUnsafeCharacters)
	}

	// Check for unsafe patterns
	for _, pattern := range unsafePatterns {
		if strings.Contains(query, pattern) {
			return fmt.Errorf("%w: unsafe pattern detected: %s", ErrUnsafeCharacters, pattern)
		}
	}

	return nil
}

// isBalancedParentheses checks if the parentheses in the string are balanced
func isBalancedParentheses(s string) bool {
	count := 0
	for _, char := range s {
		switch char {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// ValidateIdentifier checks if an identifier is safe to use with security context
func ValidateIdentifier(identifier string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating identifier: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		identifier)

	// Check identifier length
	if len(identifier) > MaxIdentifierLength {
		return fmt.Errorf("%w: identifier length %d exceeds maximum %d",
			ErrMaxLengthExceeded, len(identifier), MaxIdentifierLength)
	}

	// Check for empty identifier
	if identifier = strings.TrimSpace(identifier); identifier == "" {
		return ErrEmptyIdentifier
	}

	// Validate identifier format
	if !validIdentifierRegex.MatchString(identifier) {
		return fmt.Errorf("%w: identifier must start with a letter and contain only letters, numbers, and underscores",
			ErrInvalidIdentifier)
	}

	return nil
}

// ValidateTableName validates table name with security context
func ValidateTableName(tableName string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating table name: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		tableName)

	if tableName == "" {
		return ErrEmptyTableName
	}

	if strings.ContainsAny(tableName, "();\"'") {
		return ErrInvalidTableName
	}

	return ValidateIdentifier(tableName, ctx)
}

// ValidateFieldNames validates field names with security context
func ValidateFieldNames(fields []string, ctx *SecurityContext) error {
	for _, field := range fields {
		// Log validation attempt
		fmt.Printf("[%s] User '%s' validating field name: %s\n",
			ctx.Timestamp.Format("2006-01-02 15:04:05"),
			ctx.Username,
			field)

		if field == "" {
			return ErrEmptyFieldName
		}

		if strings.ContainsAny(field, "();\"'") {
			return ErrInvalidFieldName
		}

		if err := ValidateIdentifier(field, ctx); err != nil {
			return fmt.Errorf("invalid field '%s': %w", field, err)
		}
	}

	return nil
}

// ValidateWhereCondition validates WHERE clause with security context
func ValidateWhereCondition(where string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating WHERE condition: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		where)

	if where == "" {
		return nil
	}

	// Trim whitespace
	where = strings.TrimSpace(where)

	// Check length
	if len(where) > MaxWhereLength {
		return fmt.Errorf("%w: where condition length %d exceeds maximum %d",
			ErrMaxLengthExceeded, len(where), MaxWhereLength)
	}

	// Validate query
	if err := ValidateQuery(where, ctx); err != nil {
		return fmt.Errorf("invalid where condition: %w", err)
	}

	// Check for balanced parentheses
	if !isBalancedParentheses(where) {
		return fmt.Errorf("unbalanced parentheses in where condition")
	}

	return nil
}

// ValidateOrderBy validates ORDER BY clause with security context
func ValidateOrderBy(orderBy string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating ORDER BY: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		orderBy)

	if orderBy == "" {
		return nil
	}

	columns := strings.Split(orderBy, ",")
	for _, col := range columns {
		col = strings.TrimSpace(col)
		if strings.HasPrefix(col, "-") || strings.HasPrefix(col, "+") {
			col = col[1:]
		}
		if err := ValidateIdentifier(col, ctx); err != nil {
			return fmt.Errorf("invalid order by column '%s': %w", col, err)
		}
	}

	return nil
}

// ValidateGroupBy validates GROUP BY clause with security context
func ValidateGroupBy(groupBy string, ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating GROUP BY: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		groupBy)

	if groupBy == "" {
		return nil
	}

	columns := strings.Split(groupBy, ",")
	for _, col := range columns {
		if err := ValidateIdentifier(strings.TrimSpace(col), ctx); err != nil {
			return fmt.Errorf("invalid group by column '%s': %w", col, err)
		}
	}

	return nil
}

// ValidateDataModel validates the entire data model with security context
func ValidateDataModel[T any](dm models.DataModel[T], ctx *SecurityContext) error {
	// Log validation attempt
	fmt.Printf("[%s] User '%s' validating data model for table: %s\n",
		ctx.Timestamp.Format("2006-01-02 15:04:05"),
		ctx.Username,
		dm.TableName)

	if err := ValidateTableName(dm.TableName, ctx); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	if err := ValidateFieldNames(dm.FieldNames, ctx); err != nil {
		return fmt.Errorf("invalid field names: %w", err)
	}

	// Validate WHERE condition if present
	if dm.Where != "" {
		if err := ValidateWhereCondition(dm.Where, ctx); err != nil {
			return fmt.Errorf("invalid where condition: %w", err)
		}
	}

	// Validate ORDER BY if present
	if dm.OrderBy != "" {
		if err := ValidateOrderBy(dm.OrderBy, ctx); err != nil {
			return fmt.Errorf("invalid order by clause: %w", err)
		}
	}

	// Validate GROUP BY if present
	if dm.GroupBy != "" {
		if err := ValidateGroupBy(dm.GroupBy, ctx); err != nil {
			return fmt.Errorf("invalid group by clause: %w", err)
		}
	}

	return nil
}
