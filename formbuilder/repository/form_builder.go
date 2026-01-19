// =============================================================================
// internal/repository/form_builder_repository.go
// =============================================================================
package repository

import (
	"context"
	"time"

	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/utils"
	"p9e.in/ugcl/packages/database/sqlc"

	"github.com/google/uuid"
)

type formBuilderRepository struct {
	queries   *db.Queries
	dbManager *sqlc.DatabaseManager
}

// NewFormBuilderRepository creates a new form builder repository
func NewFormBuilderRepository(queries *db.Queries, dbManager *sqlc.DatabaseManager) IFormBuilderRepository {
	return &formBuilderRepository{
		queries:   queries,
		dbManager: dbManager,
	}
}

func (r *formBuilderRepository) CreateForm(ctx context.Context, form *db.Form) (*db.Form, error) {
	// Set timestamps and defaults
	if form.ID == uuid.Nil {
		form.ID = utils.GenerateUUID()
	}

	params := db.CreateFormParams{
		ID:                    form.ID,
		Title:                 form.Title,
		Description:           form.Description,
		Version:               form.Version,
		CreatedBy:             form.CreatedBy,
		AllowedRoles:          form.AllowedRoles,
		Audit:                 form.Audit,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		TableName:             form.TableName,
		Module:                form.Module,
		SchemaVersion:         form.SchemaVersion,
		CoreFields:            form.CoreFields,
		Steps:                 form.Steps,
		Dependencies:          form.Dependencies,
		CrossFieldValidations: form.CrossFieldValidations,
		WorkflowID:            form.WorkflowID,
	}

	return r.queries.CreateForm(ctx, params)
}

func (r *formBuilderRepository) GetForm(ctx context.Context, formID uuid.UUID) (*db.Form, error) {
	form, err := r.queries.GetForm(ctx, db.GetFormParams{ID: formID})
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (r *formBuilderRepository) GetFormByVersion(ctx context.Context, formID uuid.UUID, version string) (*db.Form, error) {
	params := db.GetFormByVersionParams{
		ID:      formID,
		Version: version,
	}

	form, err := r.queries.GetFormByVersion(ctx, params)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (r *formBuilderRepository) UpdateForm(ctx context.Context, form *db.Form) (*db.Form, error) {
	params := db.UpdateFormParams{
		ID:                    form.ID,
		Title:                 form.Title,
		Description:           form.Description,
		Version:               form.Version,
		AllowedRoles:          form.AllowedRoles,
		Audit:                 form.Audit,
		UpdatedAt:             time.Now(),
		TableName:             form.TableName,
		SchemaVersion:         form.SchemaVersion,
		CoreFields:            form.CoreFields,
		Steps:                 form.Steps,
		Dependencies:          form.Dependencies,
		CrossFieldValidations: form.CrossFieldValidations,
		WorkflowID:            form.WorkflowID,
	}

	return r.queries.UpdateForm(ctx, params)
}

func (r *formBuilderRepository) DeleteForm(ctx context.Context, formID uuid.UUID) error {
	return r.queries.DeleteForm(ctx, db.DeleteFormParams{ID: formID})
}

func (r *formBuilderRepository) ListForms(ctx context.Context, limit, offset int32) ([]*db.Form, error) {
	params := db.ListFormsParams{
		Limit:  limit,
		Offset: offset,
	}

	forms, err := r.queries.ListForms(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Form, len(forms))
	copy(result, forms)
	return result, nil
}

func (r *formBuilderRepository) CountForms(ctx context.Context) (int64, error) {
	return r.queries.CountForms(ctx)
}

func (r *formBuilderRepository) SearchForms(ctx context.Context, query string, limit, offset int32) ([]*db.Form, error) {
	params := db.SearchFormsParams{
		Column1: &query,
		Limit:   limit,
		Offset:  offset,
	}

	forms, err := r.queries.SearchForms(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Form, len(forms))
	copy(result, forms)
	return result, nil
}

func (r *formBuilderRepository) GetFormsByCreator(ctx context.Context, createdBy string, limit, offset int32) ([]*db.Form, error) {
	params := db.GetFormsByCreatorParams{
		CreatedBy: createdBy,
		Limit:     limit,
		Offset:    offset,
	}

	forms, err := r.queries.GetFormsByCreator(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*db.Form, len(forms))
	copy(result, forms)
	return result, nil
}
