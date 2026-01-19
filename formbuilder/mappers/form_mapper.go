// =============================================================================
// internal/helpers/mappers/form_mapper.go
// =============================================================================
package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/formbuilder/api/v2/form_builder"
	db "p9e.in/ugcl/formbuilder/db/generated"
)

// ProtoToFormDefinition converts protobuf FormDefinition to SQLC Form struct
func ProtoToFormDefinition(proto *pb.FormDefinition) (*db.Form, error) {
	if proto == nil || proto.Metadata == nil {
		return nil, fmt.Errorf("invalid form definition")
	}

	// Marshal JSONB fields
	stepsJSON, err := json.Marshal(proto.Steps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal steps: %w", err)
	}

	dependenciesJSON, err := json.Marshal(proto.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dependencies: %w", err)
	}

	crossValidationsJSON, err := json.Marshal(proto.CrossFieldValidations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cross validations: %w", err)
	}

	allowedRolesJSON, err := json.Marshal(proto.Metadata.AllowedRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed roles: %w", err)
	}

	coreFieldsJSON, err := json.Marshal(proto.Metadata.CoreFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal core fields: %w", err)
	}

	// Parse UUID if exists
	var formID uuid.UUID
	if proto.Metadata.FormId != "" {
		if id, err := uuid.Parse(proto.Metadata.FormId); err == nil {
			formID = id
		} else {
			formID = uuid.New()
		}
	} else {
		formID = uuid.New()
	}

	// Parse workflow ID if exists
	var workflowID uuid.NullUUID
	if proto.Workflow != nil {
		wf, err := ProtoToWorkflow(proto.Workflow)
		if err != nil {
			return nil, fmt.Errorf("failed to convert workflow: %w", err)
		}
		// service layer will persist wf and return its ID
		workflowID = uuid.NullUUID{UUID: wf.ID, Valid: true}
	}
	var tableName string
	moduleStr := proto.Metadata.Module
	tableNameStr := proto.Metadata.TableName

	// Store only the table name, not module-qualified name
	// The service layer will handle schema qualification
	if tableNameStr != "" {
		tableName = tableNameStr
	} else if moduleStr != "" {
		// If no explicit table name, use module as table name
		tableName = moduleStr
	}

	// Convert to pointers for database storage
	// Convert to pointers for database storage
	var moduleName *string
	if moduleStr != "" {
		moduleName = &moduleStr // Direct pointer creation instead of stringPtr()
	}

	return &db.Form{
		ID:                    formID,
		Title:                 proto.Metadata.Title,
		Description:           stringPtr(proto.Metadata.Description),
		Version:               proto.Metadata.Version,
		CreatedBy:             proto.Metadata.CreatedBy,
		AllowedRoles:          allowedRolesJSON,
		Audit:                 boolPtr(proto.Metadata.Audit),
		TableName:             &tableName,
		Module:                moduleName,
		SchemaVersion:         int32Ptr(proto.Metadata.SchemaVersion),
		CoreFields:            coreFieldsJSON,
		Steps:                 stepsJSON,
		Dependencies:          dependenciesJSON,
		CrossFieldValidations: crossValidationsJSON,
		WorkflowID:            workflowID,
	}, nil
}

// FormDefinitionToProto converts SQLC Form struct to protobuf FormDefinition
func FormDefinitionToProto(form *db.Form) (*pb.FormDefinition, error) {
	if form == nil {
		return nil, fmt.Errorf("invalid form model")
	}

	// Unmarshal JSONB fields
	var steps []*pb.FormStep
	if err := json.Unmarshal(form.Steps, &steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	var dependencies []*pb.Dependency
	if err := json.Unmarshal(form.Dependencies, &dependencies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
	}

	var crossValidations []*pb.CrossFieldValidation
	if err := json.Unmarshal(form.CrossFieldValidations, &crossValidations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cross validations: %w", err)
	}

	var allowedRoles []string
	if err := json.Unmarshal(form.AllowedRoles, &allowedRoles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed roles: %w", err)
	}

	var coreFields []string
	if err := json.Unmarshal(form.CoreFields, &coreFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal core fields: %w", err)
	}

	// Convert metadata
	metadata := &pb.FormMetadata{
		FormId:        form.ID.String(),
		Title:         form.Title,
		Description:   stringValue(form.Description),
		Version:       form.Version,
		CreatedBy:     form.CreatedBy,
		AllowedRoles:  allowedRoles,
		Audit:         boolValue(form.Audit),
		CreatedAt:     timestamppb.New(form.CreatedAt),
		UpdatedAt:     timestamppb.New(form.UpdatedAt),
		TableName:     stringValue(form.TableName),
		Module:        stringValue(form.Module),
		SchemaVersion: int32Value(form.SchemaVersion),
		CoreFields:    coreFields,
	}

	return &pb.FormDefinition{
		Metadata:              metadata,
		Steps:                 steps,
		Dependencies:          dependencies,
		CrossFieldValidations: crossValidations,
		// Workflow will be populated separately if needed
	}, nil
}

// Helper functions for pointer handling
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func int32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
