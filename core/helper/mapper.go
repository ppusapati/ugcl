package helper

import (
	"fmt"

	pb "p9e.in/ugcl/core/api/v2/query"
	"p9e.in/ugcl/core/models"
)

// Helper functions - reusable across all handlers
func ProtoToQueryParams(req *pb.GetQueryRequest) *models.QueryParams {
	params := &models.QueryParams{
		Page:       int(req.Page),
		Limit:      int(req.Limit),
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		Fields:     req.Fields,
		Filters:    make(map[string]interface{}),
		DateColumn: req.DateColumn,
	}

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}
	if params.DateColumn == "" {
		params.DateColumn = "created_at"
	}

	// Convert protobuf filters to map
	for _, filter := range req.Filters {
		params.Filters[filter.Field] = filter.Value
	}

	return params
}

func QueryResponseToProto(resp *models.QueryResponse) *pb.GetQueryResponse {
	var protoData []*pb.QueryRecord
	for _, record := range resp.Data {
		fields := make(map[string]string)
		for k, v := range record {
			fields[k] = fmt.Sprintf("%v", v)
		}
		protoData = append(protoData, &pb.QueryRecord{
			Fields: fields,
		})
	}

	return &pb.GetQueryResponse{
		Total: resp.Total,
		Page:  int32(resp.Page),
		Limit: int32(resp.Limit),
		Data:  protoData,
	}
}
