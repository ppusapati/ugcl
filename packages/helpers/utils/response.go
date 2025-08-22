package helpers_utils

import (
	"context"
	"errors"

	pbr "p9e.in/ugcl/packages/api/v1/response"
	"p9e.in/ugcl/packages/middleware/localize"
	"p9e.in/ugcl/packages/models"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CreateResponse constructs a generic OperationResponse.
func CreateSuccessResponse(ctx context.Context, code int32, reason, message string,
	data map[string]interface{}, id int64, uuid string) (*pbr.OperationResponse, error) {
	if message == "" { // Use reason if message is not provided
		message = reason
	}
	return &pbr.OperationResponse{
		Status: &pbr.Status{
			Code:    code,
			Reason:  reason,
			Message: localize.GetMsg(ctx, reason, message, data, nil),
		},
		Id:   wrapperspb.Int64(id),
		Uuid: wrapperspb.String(uuid),
	}, nil
}

// CreateErrorResponse constructs an error response.
func CreateErrorResponse(ctx context.Context, code int32, reason, message string, data map[string]interface{}) (*pbr.OperationResponse, error) {
	if message == "" { // Use reason if message is not provided
		message = reason
	}
	// Add error to tracing
	// tracing.AddSpanError(ctx, errors.Errorf(int(code), errorMessage, reason))

	return &pbr.OperationResponse{
		Status: &pbr.Status{
			Code:    code,
			Reason:  reason,
			Message: localize.GetMsg(ctx, reason, message, data, nil),
		},
	}, errors.New(reason)
}

func ErrorResponse[T models.Entity](ctx context.Context, message string, entity T, messageCode pbr.ErrorReason) (*pbr.OperationResponse, error) {
	return CreateErrorResponse(ctx,
		int32(messageCode),
		message,
		messageCode.String(),
		map[string]interface{}{"Entity": entity},
	)
}

// Centralized success response

func SuccessResponse[T models.Entity](ctx context.Context, message string, entity T, messageCode pbr.SuccessReason) (*pbr.OperationResponse, error) {
	return CreateSuccessResponse(ctx,
		int32(messageCode),
		message,
		messageCode.String(),
		nil, entity.GetID(), entity.GetUUID(),
	)
}

func SuccessArrayResponse[T []*models.Entity](ctx context.Context, message string, entity []*T, messageCode pbr.SuccessReason) (*pbr.OperationResponse, error) {
	return CreateSuccessResponse(ctx,
		int32(messageCode),
		message,
		messageCode.String(),
		nil, 0, "",
	)
}
