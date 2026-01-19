package services

import (
	"context"

	pb "p9e.in/ugcl/dms/api/v1"
)

// IDocumentService defines the business logic contract for document management
type IDocumentService interface {
	// Document CRUD operations
	UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error)
	GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error)
	ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error)
	UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error)
	DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error)
	GetDocumentVersions(ctx context.Context, req *pb.GetDocumentVersionsRequest) (*pb.GetDocumentVersionsResponse, error)

	// Document processing
	ProcessDocument(ctx context.Context, req *pb.ProcessDocumentRequest) (*pb.ProcessDocumentResponse, error)
	GetProcessingStatus(ctx context.Context, req *pb.GetProcessingStatusRequest) (*pb.GetProcessingStatusResponse, error)
	GenerateThumbnail(ctx context.Context, req *pb.GenerateThumbnailRequest) (*pb.GenerateThumbnailResponse, error)
	ExtractText(ctx context.Context, req *pb.ExtractTextRequest) (*pb.ExtractTextResponse, error)

	// Document viewing
	GetDocumentPreview(ctx context.Context, req *pb.GetDocumentPreviewRequest) (*pb.GetDocumentPreviewResponse, error)
	GetDocumentPage(ctx context.Context, req *pb.GetDocumentPageRequest) (*pb.GetDocumentPageResponse, error)
	StreamDocument(ctx context.Context, req *pb.StreamDocumentRequest, stream pb.DMSService_StreamDocumentServer) error
}

// IDocumentShareService defines the business logic contract for document sharing
type IDocumentShareService interface {
	CreateDocumentShare(ctx context.Context, req *pb.CreateDocumentShareRequest) (*pb.CreateDocumentShareResponse, error)
	GetDocumentShare(ctx context.Context, req *pb.GetDocumentShareRequest) (*pb.GetDocumentShareResponse, error)
	ListDocumentShares(ctx context.Context, req *pb.ListDocumentSharesRequest) (*pb.ListDocumentSharesResponse, error)
	GetUserShares(ctx context.Context, req *pb.GetUserSharesRequest) (*pb.GetUserSharesResponse, error)
	DeleteDocumentShare(ctx context.Context, req *pb.DeleteDocumentShareRequest) (*pb.DeleteDocumentShareResponse, error)
}

// IWatermarkService defines the business logic contract for watermark management
type IWatermarkService interface {
	CreateWatermarkConfig(ctx context.Context, req *pb.CreateWatermarkConfigRequest) (*pb.CreateWatermarkConfigResponse, error)
	GetWatermarkConfigs(ctx context.Context, req *pb.GetWatermarkConfigsRequest) (*pb.GetWatermarkConfigsResponse, error)
	UpdateWatermarkConfig(ctx context.Context, req *pb.UpdateWatermarkConfigRequest) (*pb.UpdateWatermarkConfigResponse, error)
	DeleteWatermarkConfig(ctx context.Context, req *pb.DeleteWatermarkConfigRequest) (*pb.DeleteWatermarkConfigResponse, error)
	ApplyWatermark(ctx context.Context, req *pb.ApplyWatermarkRequest) (*pb.ApplyWatermarkResponse, error)
}

// IAnalyticsService defines the business logic contract for DMS analytics
type IAnalyticsService interface {
	GetDocumentAccess(ctx context.Context, req *pb.GetDocumentAccessRequest) (*pb.GetDocumentAccessResponse, error)
	GetDocumentAnalytics(ctx context.Context, req *pb.GetDocumentAnalyticsRequest) (*pb.GetDocumentAnalyticsResponse, error)
	GetStorageUsage(ctx context.Context, req *pb.GetStorageUsageRequest) (*pb.GetStorageUsageResponse, error)
}
