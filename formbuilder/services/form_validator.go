package services

import (
	"fmt"
	"net/url"
	"strings"

	formbuilderv2 "p9e.in/ugcl/formbuilder/api/v2/form_builder"

	"google.golang.org/protobuf/types/known/anypb"
)

// FormValidator handles form validation
type FormValidator struct {
	validators map[string]ValidatorFunc
}

type ValidatorFunc func(value interface{}, field *formbuilderv2.FormField) error

func NewFormValidator() *FormValidator {
	fv := &FormValidator{
		validators: make(map[string]ValidatorFunc),
	}
	fv.registerDefaultValidators()
	return fv
}

// registerDefaultValidators registers built-in validators
func (fv *FormValidator) registerDefaultValidators() {
	// Email validator
	fv.validators["email"] = func(value interface{}, field *formbuilderv2.FormField) error {
		// Simple email validation
		str, ok := value.(string)
		if !ok || !strings.Contains(str, "@") {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}

	// URL validator
	fv.validators["url"] = func(value interface{}, field *formbuilderv2.FormField) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid URL format")
		}
		_, err := url.Parse(str)
		return err
	}
}

// ValidateForm validates form instance data
func (fv *FormValidator) ValidateForm(formDef *formbuilderv2.FormDefinition, data map[string]*anypb.Any) []error {
	var errors []error

	// Field-level validation
	for _, step := range formDef.Steps {
		for _, field := range step.Fields {
			if err := fv.validateField(field, data[field.Id]); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Cross-field validation
	for _, cfv := range formDef.CrossFieldValidations {
		if err := fv.validateCrossField(cfv, data); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateField validates a single field
func (fv *FormValidator) validateField(field *formbuilderv2.FormField, value *anypb.Any) error {
	// Check required
	if field.Required && (value == nil || isEmptyValue(value)) {
		return fmt.Errorf("field %s is required", field.Id)
	}

	// Skip validation if value is empty and not required
	if !field.Required && (value == nil || isEmptyValue(value)) {
		return nil
	}

	// Type-specific validation
	switch field.Type {
	case formbuilderv2.FieldType_EMAIL:
		return fv.validators["email"](value, field)
	case formbuilderv2.FieldType_URL:
		return fv.validators["url"](value, field)
	case formbuilderv2.FieldType_NUMBER:
		return fv.validateNumber(value, field)
	}

	return nil
}

// validateNumber validates numeric fields
func (fv *FormValidator) validateNumber(value *anypb.Any, field *formbuilderv2.FormField) error {
	if field.Validation == nil {
		return nil
	}

	// Extract numeric value
	// In real implementation, properly unmarshal the Any value
	// For now, assume it's a float64
	num := 0.0 // Extract from value

	if field.Validation.Min > 0 && num < field.Validation.Min {
		return fmt.Errorf("value must be at least %f", field.Validation.Min)
	}

	if field.Validation.Max > 0 && num > field.Validation.Max {
		return fmt.Errorf("value must be at most %f", field.Validation.Max)
	}

	return nil
}

// validateCrossField validates cross-field dependencies
func (fv *FormValidator) validateCrossField(cfv *formbuilderv2.CrossFieldValidation, data map[string]*anypb.Any) error {
	// Implementation would evaluate the rule expression
	// This could use an expression evaluation library
	return nil
}

// Helper function
func isEmptyValue(value *anypb.Any) bool {
	if value == nil {
		return true
	}
	// Check for empty string, zero values, etc.
	// Implementation depends on how you marshal values
	return false
}
