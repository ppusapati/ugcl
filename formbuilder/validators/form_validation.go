// =============================================================================
// internal/helpers/validators/form_validator.go
// =============================================================================
package validators

import (
	"encoding/json"
	"fmt"
	"regexp"

	"p9e.in/ugcl/formbuilder/api/v2/form_builder"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// FormStep represents a form step for validation
type FormStep struct {
	ID     string      `json:"id"`
	Label  string      `json:"label"`
	Fields []FormField `json:"fields"`
	Order  int32       `json:"order"`
}

// FormField represents a form field for validation
type FormField struct {
	ID         string                 `json:"id"`
	Type       form_builder.FieldType `json:"type"`
	Label      string                 `json:"label"`
	Required   bool                   `json:"required"`
	Validation FieldValidation        `json:"validation"`
}

// FieldValidation represents field validation rules
type FieldValidation struct {
	MinLength     *int32   `json:"min_length,omitempty"`
	MaxLength     *int32   `json:"max_length,omitempty"`
	Pattern       string   `json:"pattern,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
}

// ValidateForm validates a complete SQLC form struct
func ValidateForm(form *db.Form) []ValidationError {
	var errors []ValidationError

	// Basic validation
	if form.Title == "" {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Title is required",
			Code:    "REQUIRED",
		})
	}

	// Parse and validate steps
	var steps []FormStep
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		errors = append(errors, ValidationError{
			Field:   "steps",
			Message: "Invalid steps format",
			Code:    "INVALID_FORMAT",
		})
		return errors
	}

	if len(steps) == 0 {
		errors = append(errors, ValidationError{
			Field:   "steps",
			Message: "At least one step is required",
			Code:    "REQUIRED",
		})
	}

	// Validate each step
	for i, step := range steps {
		stepErrors := ValidateFormStep(&step, fmt.Sprintf("steps[%d]", i))
		errors = append(errors, stepErrors...)
	}

	// Validate version format
	if form.Version != "" {
		matched, err := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, form.Version)
		if err != nil || !matched {
			errors = append(errors, ValidationError{
				Field:   "version",
				Message: "Version must be in semver format (e.g., 1.0.0)",
				Code:    "INVALID_FORMAT",
			})
		}
	}

	return errors
}

// ValidateFormStep validates a form step
func ValidateFormStep(step *FormStep, fieldPrefix string) []ValidationError {
	var errors []ValidationError

	if step.ID == "" {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.id", fieldPrefix),
			Message: "Step ID is required",
			Code:    "REQUIRED",
		})
	}

	if step.Label == "" {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.label", fieldPrefix),
			Message: "Step label is required",
			Code:    "REQUIRED",
		})
	}

	if len(step.Fields) == 0 {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.fields", fieldPrefix),
			Message: "At least one field is required per step",
			Code:    "REQUIRED",
		})
	}

	// Validate each field
	for j, field := range step.Fields {
		fieldErrors := ValidateFormField(&field, fmt.Sprintf("%s.fields[%d]", fieldPrefix, j))
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// ValidateFormField validates a form field definition
func ValidateFormField(field *FormField, fieldPrefix string) []ValidationError {
	var errors []ValidationError

	if field.ID == "" {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.id", fieldPrefix),
			Message: "Field ID is required",
			Code:    "REQUIRED",
		})
	}

	if field.Label == "" {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.label", fieldPrefix),
			Message: "Field label is required",
			Code:    "REQUIRED",
		})
	}

	// Validate field type
	if !isValidFieldType(field.Type) {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.type", fieldPrefix),
			Message: "Invalid field type",
			Code:    "INVALID_TYPE",
		})
	}

	// Validate field validation rules
	if field.Validation.Pattern != "" {
		if _, err := regexp.Compile(field.Validation.Pattern); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.validation.pattern", fieldPrefix),
				Message: "Invalid regex pattern",
				Code:    "INVALID_PATTERN",
			})
		}
	}

	// Validate length constraints
	if field.Validation.MinLength != nil && field.Validation.MaxLength != nil {
		if *field.Validation.MinLength > *field.Validation.MaxLength {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.validation", fieldPrefix),
				Message: "Min length cannot be greater than max length",
				Code:    "INVALID_RANGE",
			})
		}
	}

	return errors
}

// ValidateFieldValue validates a field value against field definition
func ValidateFieldValue(field *FormField, value interface{}) []ValidationError {
	var errors []ValidationError

	// Check required fields
	if field.Required && (value == nil || value == "") {
		errors = append(errors, ValidationError{
			Field:   field.ID,
			Message: fmt.Sprintf("%s is required", field.Label),
			Code:    "REQUIRED",
		})
		return errors
	}

	if value == nil {
		return errors
	}

	// Convert value to string for common validations
	strValue := fmt.Sprintf("%v", value)

	// Validate length constraints
	if field.Validation.MinLength != nil && len(strValue) < int(*field.Validation.MinLength) {
		errors = append(errors, ValidationError{
			Field:   field.ID,
			Message: fmt.Sprintf("%s must be at least %d characters long", field.Label, *field.Validation.MinLength),
			Code:    "MIN_LENGTH",
		})
	}

	if field.Validation.MaxLength != nil && len(strValue) > int(*field.Validation.MaxLength) {
		errors = append(errors, ValidationError{
			Field:   field.ID,
			Message: fmt.Sprintf("%s must be at most %d characters long", field.Label, *field.Validation.MaxLength),
			Code:    "MAX_LENGTH",
		})
	}

	// Validate pattern
	if field.Validation.Pattern != "" {
		matched, err := regexp.MatchString(field.Validation.Pattern, strValue)
		if err != nil || !matched {
			errors = append(errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("%s format is invalid", field.Label),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	// Validate allowed values
	if len(field.Validation.AllowedValues) > 0 {
		found := false
		for _, allowedValue := range field.Validation.AllowedValues {
			if strValue == allowedValue {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("%s has an invalid value", field.Label),
				Code:    "INVALID_VALUE",
			})
		}
	}

	// Type-specific validation
	switch field.Type {
	case form_builder.FieldType_EMAIL:
		if !isValidEmail(strValue) {
			errors = append(errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("%s must be a valid email address", field.Label),
				Code:    "INVALID_EMAIL",
			})
		}
	case form_builder.FieldType_URL:
		if !isValidURL(strValue) {
			errors = append(errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("%s must be a valid URL", field.Label),
				Code:    "INVALID_URL",
			})
		}
	case form_builder.FieldType_PHONE:
		if !isValidPhone(strValue) {
			errors = append(errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("%s must be a valid phone number", field.Label),
				Code:    "INVALID_PHONE",
			})
		}
	}

	return errors
}

func isValidFieldType(fieldType form_builder.FieldType) bool {
	switch fieldType {
	case form_builder.FieldType_TEXT,
		form_builder.FieldType_NUMBER,
		form_builder.FieldType_EMAIL,
		form_builder.FieldType_DROPDOWN,
		form_builder.FieldType_RADIO,
		form_builder.FieldType_CHECKBOX,
		form_builder.FieldType_DATE,
		form_builder.FieldType_DATETIME,
		form_builder.FieldType_FILE,
		form_builder.FieldType_TEXTAREA,
		form_builder.FieldType_MULTI_SELECT,
		form_builder.FieldType_CURRENCY,
		form_builder.FieldType_PHONE,
		form_builder.FieldType_URL,
		form_builder.FieldType_JSON,
		form_builder.FieldType_ARRAY,
		form_builder.FieldType_NESTED_FORM:
		return true
	default:
		return false
	}
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

func isValidURL(url string) bool {
	urlRegex := `^https?://[^\s/$.?#].[^\s]*`
	matched, _ := regexp.MatchString(urlRegex, url)
	return matched
}

func isValidPhone(phone string) bool {
	phoneRegex := `^\+?[\d\s\-\(\)]{10,}`
	matched, _ := regexp.MatchString(phoneRegex, phone)
	return matched
}

// ValidateFormInstanceFieldValues validates field values for a form instance
func ValidateFormInstanceFieldValues(form *db.Form, fieldValues map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Parse form steps
	var steps []FormStep
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		errors = append(errors, ValidationError{
			Field:   "form",
			Message: "Invalid form definition",
			Code:    "INVALID_FORM",
		})
		return errors
	}

	// Validate each field
	for _, step := range steps {
		for _, field := range step.Fields {
			value, exists := fieldValues[field.ID]

			// Skip validation if field doesn't exist and is not required
			if !exists && !field.Required {
				continue
			}

			fieldErrors := ValidateFieldValue(&field, value)
			errors = append(errors, fieldErrors...)
		}
	}

	return errors
}
