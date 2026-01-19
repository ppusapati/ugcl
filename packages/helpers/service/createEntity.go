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
)

func CreateEntity[T models.Entity, P proto.Message](
	ctx context.Context,
	req P,
	convertFunc func(P) T, // Converts Protobuf to DB model
	tracer *tracing.TracingProvider, // Tracing provider
	metrics metrics.MetricsProvider, // Metrics provider
	repoFunc func(context.Context, T) (T, error), // Repository function
) (*pbr.OperationResponse, error) {
	var zeroValue T                   // Declare zero value once
	entityType := hu.GetTypeName[T]() // Get entity type name

	// Start tracing
	ctx, span := tracer.StartSpan(ctx, "Create"+entityType)
	defer span.End()
	tracer.AddSpanTags(ctx, map[string]string{"operation": entityType + "_create"})

	// Validate request
	if err := hu.ValidateProto(req); err != nil {
		return hu.ErrorResponse(ctx, entityType+" validation failed", zeroValue, pbr.ErrorReason_VALIDATION_FAILED)
	}

	// Convert request to DB model & track operation time
	startTime := time.Now()
	result, err := repoFunc(ctx, convertFunc(req))
	recordMetric(metrics, entityType, "Create", startTime, err == nil)

	if err != nil {
		return hu.ErrorResponse(ctx, "Failed to create "+entityType, zeroValue, pbr.ErrorReason_CREATION_FAILED)
	}

	return hu.SuccessResponse(ctx, entityType+" created successfully", result, pbr.SuccessReason_CREATED_SUCCESSFULLY)
}
