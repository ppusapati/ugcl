package helpers_service

import (
	"context"
	"time"

	pbr "p9e.in/ugcl/packages/api/v1/response"
	hu "p9e.in/ugcl/packages/helpers/utils"
	"p9e.in/ugcl/packages/metrics"
	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/tracing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Generic UpdateEntity function
func UpdateEntity[T models.Entity, P proto.Message](
	ctx context.Context,
	req P,
	id int64,
	uuid string,
	updateMask *fieldmaskpb.FieldMask,
	convertFunc func(P) T, // Converts Protobuf to DB model
	toProtoFunc func(T) P, // Converts DB model to Protobuf
	tracer *tracing.TracingProvider,
	metrics metrics.MetricsProvider,
	repoGetByIDFunc func(context.Context, int64) (T, error),
	repoGetByUUIDFunc func(context.Context, string) (T, error),
	repoUpdateFunc func(context.Context, hu.UpdateRequest[T]) (T, error),
) (*pbr.OperationResponse, error) {
	var zeroValue T                   // Declare zero value once
	entityType := hu.GetTypeName[T]() // Get entity name

	// Start tracing
	ctx, span := tracer.StartSpan(ctx, "Update"+entityType)
	defer span.End()
	tracer.AddSpanTags(ctx, map[string]string{"operation": entityType + "_update"})

	// Validate request
	if err := hu.ValidateProto(req); err != nil {
		return hu.ErrorResponse(ctx, entityType+" validation failed", zeroValue, pbr.ErrorReason_VALIDATION_FAILED)
	}

	// Retrieve existing entity
	existingEntity, err := fetchEntity(ctx, id, uuid, repoGetByIDFunc, repoGetByUUIDFunc)
	if err != nil {
		return hu.ErrorResponse(ctx, entityType+" not found", zeroValue, pbr.ErrorReason_NOT_FOUND)
	}

	// Apply field mask if provided
	if updateMask != nil {
		hu.ApplyFieldMask(updateMask, toProtoFunc(existingEntity), req)
	}

	var fieldMaskPaths []string
	if updateMask != nil {
		fieldMaskPaths = updateMask.Paths
	}
	// Convert updated Protobuf model back to DB model
	dbModel := convertFunc(req)

	// Create update request with field mask
	updateReq := hu.UpdateRequest[T]{
		Entity:    dbModel,
		FieldMask: fieldMaskPaths,
		UpdatedBy: "admin", //TODO: getUserID(ctx),
		UpdatedAt: time.Date(2025, 3, 21, 11, 33, 23, 0, time.UTC),
	}
	// Perform update operation with metrics
	startTime := time.Now()
	updatedEntity, err := repoUpdateFunc(ctx, updateReq)
	recordMetric(metrics, entityType, "Update", startTime, err == nil)

	if err != nil {
		return hu.ErrorResponse(ctx, "Failed to update "+entityType, zeroValue, pbr.ErrorReason_UPDATE_FAILED)
	}

	return hu.SuccessResponse(ctx, entityType+" updated successfully", updatedEntity, pbr.SuccessReason_UPDATED_SUCCESSFULLY)
}
