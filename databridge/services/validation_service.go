// =============================================================================
// services/validation_service.go - Data validation service implementation
// =============================================================================
package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/models"

	"github.com/google/uuid"
)

// ValidationService implements IValidationService
type ValidationService struct{}

// NewValidationService creates a new validation service instance
func NewValidationService() IValidationService {
	return &ValidationService{}
}

// ValidateFieldValue validates a single field value against column constraints
func (s *ValidationService) ValidateFieldValue(value string, column *pb.Column) error {
	value = strings.TrimSpace(value)

	// Check required constraint
	if column.IsRequired && value == "" {
		return fmt.Errorf("field '%s' is required but empty", column.DisplayName)
	}

	// Skip further validation if value is empty and not required
	if value == "" {
		return nil
	}

	// Validate data type compatibility
	if err := s.validateDataType(value, column.DataType); err != nil {
		return fmt.Errorf("field '%s': %w", column.DisplayName, err)
	}

	// Check max length constraint
	if column.MaxLength != nil && int32(len(value)) > *column.MaxLength {
		return fmt.Errorf("field '%s' exceeds maximum length of %d characters", column.DisplayName, *column.MaxLength)
	}

	// Apply validation rules
	for _, rule := range column.ValidationRules {
		if err := s.applyValidationRule(value, rule, column.DisplayName); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRequiredFields checks that all required fields are present and not empty
func (s *ValidationService) ValidateRequiredFields(data map[string]string, requiredColumns []*pb.Column) []string {
	var missingFields []string

	for _, column := range requiredColumns {
		if !column.IsRequired {
			continue
		}

		// Find the mapped CSV field for this column
		value, exists := data[column.ColumnName]
		if !exists || strings.TrimSpace(value) == "" {
			missingFields = append(missingFields, column.DisplayName)
		}
	}

	return missingFields
}

// TransformFieldValue applies transformation to a field value
func (s *ValidationService) TransformFieldValue(value, transformation string, dataType pb.DataType) (interface{}, error) {
	value = strings.TrimSpace(value)

	// Apply transformation
	switch transformation {
	case "trim":
		value = strings.TrimSpace(value)
	case "uppercase":
		value = strings.ToUpper(value)
	case "lowercase":
		value = strings.ToLower(value)
	case "title_case":
		value = strings.Title(strings.ToLower(value))
	}

	// Convert to target data type
	return s.convertToDataType(value, dataType)
}

// ValidateDataTypes validates data types for multiple fields
func (s *ValidationService) ValidateDataTypes(data map[string]string, columns []*pb.Column) []*models.ValidationError {
	var errors []*models.ValidationError

	for _, column := range columns {
		value, exists := data[column.ColumnName]
		if !exists {
			continue
		}

		if err := s.ValidateFieldValue(value, column); err != nil {
			errors = append(errors, &models.ValidationError{
				FieldName:    column.ColumnName,
				ColumnName:   column.ColumnName,
				Value:        value,
				ErrorType:    "validation",
				ErrorMessage: err.Error(),
				Severity:     models.ValidationLevelError,
			})
		}
	}

	return errors
}

// ApplyBusinessRules applies custom business validation rules
func (s *ValidationService) ApplyBusinessRules(data map[string]string, rules []*pb.ValidationRule) []*models.ValidationError {
	var errors []*models.ValidationError

	for _, rule := range rules {
		if err := s.evaluateBusinessRule(data, rule); err != nil {
			errors = append(errors, &models.ValidationError{
				ErrorType:    "business_rule",
				ErrorMessage: err.Error(),
				Severity:     models.ValidationLevelError,
			})
		}
	}

	return errors
}

// Helper methods

func (s *ValidationService) validateDataType(value string, dataType pb.DataType) error {
	switch dataType {
	case pb.DataType_DATA_TYPE_INTEGER, pb.DataType_DATA_TYPE_BIGINT, pb.DataType_DATA_TYPE_SMALLINT:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}

	case pb.DataType_DATA_TYPE_DECIMAL, pb.DataType_DATA_TYPE_NUMERIC, pb.DataType_DATA_TYPE_REAL, pb.DataType_DATA_TYPE_DOUBLE_PRECISION:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid decimal value: %s", value)
		}

	case pb.DataType_DATA_TYPE_BOOLEAN:
		lower := strings.ToLower(value)
		if lower != "true" && lower != "false" && lower != "1" && lower != "0" {
			return fmt.Errorf("invalid boolean value: %s (expected: true, false, 1, or 0)", value)
		}

	case pb.DataType_DATA_TYPE_UUID:
		if _, err := uuid.Parse(value); err != nil {
			return fmt.Errorf("invalid UUID value: %s", value)
		}

	case pb.DataType_DATA_TYPE_DATE:
		dateFormats := []string{
			"2006-01-02",
			"01/02/2006",
			"02-01-2006",
			"2006/01/02",
		}
		valid := false
		for _, format := range dateFormats {
			if _, err := time.Parse(format, value); err == nil {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid date value: %s (expected formats: YYYY-MM-DD, MM/DD/YYYY, etc.)", value)
		}

	case pb.DataType_DATA_TYPE_TIME:
		timeFormats := []string{
			"15:04:05",
			"15:04",
			"3:04 PM",
			"3:04:05 PM",
		}
		valid := false
		for _, format := range timeFormats {
			if _, err := time.Parse(format, value); err == nil {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid time value: %s", value)
		}

	case pb.DataType_DATA_TYPE_TIMESTAMP, pb.DataType_DATA_TYPE_TIMESTAMPTZ:
		timestampFormats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
			"01/02/2006 15:04:05",
			"2006-01-02 15:04:05.000",
		}
		valid := false
		for _, format := range timestampFormats {
			if _, err := time.Parse(format, value); err == nil {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid timestamp value: %s", value)
		}

	case pb.DataType_DATA_TYPE_JSON, pb.DataType_DATA_TYPE_JSONB:
		// Basic JSON validation - check if it starts and ends with { } or [ ]
		trimmed := strings.TrimSpace(value)
		if !((strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))) {
			return fmt.Errorf("invalid JSON value: %s", value)
		}

	case pb.DataType_DATA_TYPE_TEXT, pb.DataType_DATA_TYPE_VARCHAR, pb.DataType_DATA_TYPE_CHAR:
		// Text types don't need special validation beyond length checks
		break

	default:
		// Unknown data type, treat as text
		break
	}

	return nil
}

func (s *ValidationService) applyValidationRule(value string, rule *pb.ValidationRule, fieldName string) error {
	switch rule.Type {
	case "min_length":
		minLen, err := strconv.Atoi(rule.Value)
		if err != nil {
			return fmt.Errorf("invalid min_length rule value")
		}
		if len(value) < minLen {
			return fmt.Errorf("field '%s' must be at least %d characters", fieldName, minLen)
		}

	case "max_length":
		maxLen, err := strconv.Atoi(rule.Value)
		if err != nil {
			return fmt.Errorf("invalid max_length rule value")
		}
		if len(value) > maxLen {
			return fmt.Errorf("field '%s' must be at most %d characters", fieldName, maxLen)
		}

	case "pattern":
		matched, err := regexp.MatchString(rule.Value, value)
		if err != nil {
			return fmt.Errorf("invalid regex pattern in validation rule")
		}
		if !matched {
			message := rule.Message
			if message == "" {
				message = fmt.Sprintf("field '%s' does not match required pattern", fieldName)
			}
			return fmt.Errorf(message)
		}

	case "email":
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, _ := regexp.MatchString(emailPattern, value)
		if !matched {
			return fmt.Errorf("field '%s' must be a valid email address", fieldName)
		}

	case "phone":
		// Basic phone number pattern
		phonePattern := `^\+?[\d\s\-\(\)]{10,}$`
		matched, _ := regexp.MatchString(phonePattern, value)
		if !matched {
			return fmt.Errorf("field '%s' must be a valid phone number", fieldName)
		}

	case "url":
		// Basic URL pattern
		urlPattern := `^https?://[^\s/$.?#].[^\s]*$`
		matched, _ := regexp.MatchString(urlPattern, value)
		if !matched {
			return fmt.Errorf("field '%s' must be a valid URL", fieldName)
		}

	case "min_value":
		numericValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("field '%s' must be numeric for min_value validation", fieldName)
		}
		minValue, err := strconv.ParseFloat(rule.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid min_value rule value")
		}
		if numericValue < minValue {
			return fmt.Errorf("field '%s' must be at least %g", fieldName, minValue)
		}

	case "max_value":
		numericValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("field '%s' must be numeric for max_value validation", fieldName)
		}
		maxValue, err := strconv.ParseFloat(rule.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid max_value rule value")
		}
		if numericValue > maxValue {
			return fmt.Errorf("field '%s' must be at most %g", fieldName, maxValue)
		}

	case "in_list":
		allowedValues := strings.Split(rule.Value, ",")
		found := false
		for _, allowed := range allowedValues {
			if strings.TrimSpace(allowed) == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field '%s' must be one of: %s", fieldName, rule.Value)
		}

	case "not_in_list":
		forbiddenValues := strings.Split(rule.Value, ",")
		for _, forbidden := range forbiddenValues {
			if strings.TrimSpace(forbidden) == value {
				return fmt.Errorf("field '%s' cannot be: %s", fieldName, value)
			}
		}

	default:
		// Unknown validation rule type, skip
		break
	}

	return nil
}

func (s *ValidationService) convertToDataType(value string, dataType pb.DataType) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	switch dataType {
	case pb.DataType_DATA_TYPE_INTEGER, pb.DataType_DATA_TYPE_BIGINT:
		return strconv.ParseInt(value, 10, 64)

	case pb.DataType_DATA_TYPE_SMALLINT:
		val, err := strconv.ParseInt(value, 10, 16)
		return int16(val), err

	case pb.DataType_DATA_TYPE_DECIMAL, pb.DataType_DATA_TYPE_NUMERIC, pb.DataType_DATA_TYPE_REAL, pb.DataType_DATA_TYPE_DOUBLE_PRECISION:
		return strconv.ParseFloat(value, 64)

	case pb.DataType_DATA_TYPE_BOOLEAN:
		lower := strings.ToLower(value)
		if lower == "true" || lower == "1" {
			return true, nil
		} else if lower == "false" || lower == "0" {
			return false, nil
		}
		return strconv.ParseBool(value)

	case pb.DataType_DATA_TYPE_UUID:
		parsed, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		return parsed.String(), nil

	case pb.DataType_DATA_TYPE_DATE:
		dateFormats := []string{
			"2006-01-02",
			"01/02/2006",
			"02-01-2006",
			"2006/01/02",
		}
		for _, format := range dateFormats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
		return nil, fmt.Errorf("invalid date format")

	case pb.DataType_DATA_TYPE_TIME:
		timeFormats := []string{
			"15:04:05",
			"15:04",
			"3:04 PM",
			"3:04:05 PM",
		}
		for _, format := range timeFormats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Format("15:04:05"), nil
			}
		}
		return nil, fmt.Errorf("invalid time format")

	case pb.DataType_DATA_TYPE_TIMESTAMP, pb.DataType_DATA_TYPE_TIMESTAMPTZ:
		timestampFormats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
			"01/02/2006 15:04:05",
			"2006-01-02 15:04:05.000",
		}
		for _, format := range timestampFormats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Format("2006-01-02 15:04:05"), nil
			}
		}
		return nil, fmt.Errorf("invalid timestamp format")

	default:
		// For text types and unknown types, return as string
		return value, nil
	}
}

func (s *ValidationService) evaluateBusinessRule(data map[string]string, rule *pb.ValidationRule) error {
	// This is a simplified implementation
	// In a real system, you might want to use a rule engine or expression evaluator

	switch rule.Type {
	case "conditional_required":
		// Example: if field A has value X, then field B is required
		// Rule format: "fieldA:valueA|fieldB"
		parts := strings.Split(rule.Value, "|")
		if len(parts) != 2 {
			return fmt.Errorf("invalid conditional_required rule format")
		}

		conditionParts := strings.Split(parts[0], ":")
		if len(conditionParts) != 2 {
			return fmt.Errorf("invalid conditional_required condition format")
		}

		conditionField := conditionParts[0]
		conditionValue := conditionParts[1]
		requiredField := parts[1]

		if data[conditionField] == conditionValue {
			if value := data[requiredField]; value == "" {
				message := rule.Message
				if message == "" {
					message = fmt.Sprintf("field '%s' is required when '%s' is '%s'", requiredField, conditionField, conditionValue)
				}
				return fmt.Errorf(message)
			}
		}

	case "mutual_exclusive":
		// Example: only one of field A or field B can have a value
		fields := strings.Split(rule.Value, ",")
		filledFields := 0
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if value := data[field]; value != "" {
				filledFields++
			}
		}

		if filledFields > 1 {
			message := rule.Message
			if message == "" {
				message = fmt.Sprintf("only one of these fields can have a value: %s", rule.Value)
			}
			return fmt.Errorf(message)
		}

	case "at_least_one":
		// Example: at least one of the specified fields must have a value
		fields := strings.Split(rule.Value, ",")
		hasValue := false
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if value := data[field]; value != "" {
				hasValue = true
				break
			}
		}

		if !hasValue {
			message := rule.Message
			if message == "" {
				message = fmt.Sprintf("at least one of these fields must have a value: %s", rule.Value)
			}
			return fmt.Errorf(message)
		}

	default:
		// Unknown business rule type, skip
		break
	}

	return nil
}