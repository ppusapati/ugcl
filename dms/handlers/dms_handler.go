package handlers

import (
	"context"

	"connectrpc.com/connect"

	pb "p9e.in/ugcl/dms/api/v1"
	"p9e.in/ugcl/dms/api/v1/dmsv1connect"
	"p9e.in/ugcl/dms/services"
)

// DMSHandler implements the Connect RPC handlers for DMSService
type DMSHandler struct {
	documentService  services.IDocumentService
	shareService     services.IDocumentShareService
	watermarkService services.IWatermarkService
	analyticsService services.IAnalyticsService
}

// NewDMSHandler creates a new DMS handler
func NewDMSHandler(
	documentService services.IDocumentService,
	shareService services.IDocumentShareService,
	watermarkService services.IWatermarkService,
	analyticsService services.IAnalyticsService,
) dmsv1connect.DMSServiceHandler {
	return &DMSHandler{
		documentService:  documentService,
		shareService:     shareService,
		watermarkService: watermarkService,
		analyticsService: analyticsService,
	}
}

// Document CRUD handlers

func (h *DMSHandler) UploadDocument(
	ctx context.Context,
	req *connect.Request[pb.UploadDocumentRequest],
) (*connect.Response[pb.UploadDocumentResponse], error) {
	response, err := h.documentService.UploadDocument(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetDocument(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentRequest],
) (*connect.Response[pb.GetDocumentResponse], error) {
	response, err := h.documentService.GetDocument(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) ListDocuments(
	ctx context.Context,
	req *connect.Request[pb.ListDocumentsRequest],
) (*connect.Response[pb.ListDocumentsResponse], error) {
	response, err := h.documentService.ListDocuments(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) UpdateDocument(
	ctx context.Context,
	req *connect.Request[pb.UpdateDocumentRequest],
) (*connect.Response[pb.UpdateDocumentResponse], error) {
	response, err := h.documentService.UpdateDocument(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) DeleteDocument(
	ctx context.Context,
	req *connect.Request[pb.DeleteDocumentRequest],
) (*connect.Response[pb.DeleteDocumentResponse], error) {
	response, err := h.documentService.DeleteDocument(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetDocumentVersions(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentVersionsRequest],
) (*connect.Response[pb.GetDocumentVersionsResponse], error) {
	response, err := h.documentService.GetDocumentVersions(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

// Document processing handlers

func (h *DMSHandler) ProcessDocument(
	ctx context.Context,
	req *connect.Request[pb.ProcessDocumentRequest],
) (*connect.Response[pb.ProcessDocumentResponse], error) {
	response, err := h.documentService.ProcessDocument(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetProcessingStatus(
	ctx context.Context,
	req *connect.Request[pb.GetProcessingStatusRequest],
) (*connect.Response[pb.GetProcessingStatusResponse], error) {
	response, err := h.documentService.GetProcessingStatus(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GenerateThumbnail(
	ctx context.Context,
	req *connect.Request[pb.GenerateThumbnailRequest],
) (*connect.Response[pb.GenerateThumbnailResponse], error) {
	response, err := h.documentService.GenerateThumbnail(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) ExtractText(
	ctx context.Context,
	req *connect.Request[pb.ExtractTextRequest],
) (*connect.Response[pb.ExtractTextResponse], error) {
	response, err := h.documentService.ExtractText(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

// Document viewing handlers

func (h *DMSHandler) GetDocumentPreview(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentPreviewRequest],
) (*connect.Response[pb.GetDocumentPreviewResponse], error) {
	response, err := h.documentService.GetDocumentPreview(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetDocumentPage(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentPageRequest],
) (*connect.Response[pb.GetDocumentPageResponse], error) {
	response, err := h.documentService.GetDocumentPage(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) StreamDocument(
	ctx context.Context,
	req *connect.Request[pb.StreamDocumentRequest],
	stream *connect.ServerStream[pb.StreamDocumentResponse],
) error {
	// Streaming handler delegates to service
	return h.documentService.StreamDocument(ctx, req.Msg, stream)
}

// Watermark handlers

func (h *DMSHandler) CreateWatermarkConfig(
	ctx context.Context,
	req *connect.Request[pb.CreateWatermarkConfigRequest],
) (*connect.Response[pb.CreateWatermarkConfigResponse], error) {
	response, err := h.watermarkService.CreateWatermarkConfig(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetWatermarkConfigs(
	ctx context.Context,
	req *connect.Request[pb.GetWatermarkConfigsRequest],
) (*connect.Response[pb.GetWatermarkConfigsResponse], error) {
	response, err := h.watermarkService.GetWatermarkConfigs(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) UpdateWatermarkConfig(
	ctx context.Context,
	req *connect.Request[pb.UpdateWatermarkConfigRequest],
) (*connect.Response[pb.UpdateWatermarkConfigResponse], error) {
	response, err := h.watermarkService.UpdateWatermarkConfig(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) DeleteWatermarkConfig(
	ctx context.Context,
	req *connect.Request[pb.DeleteWatermarkConfigRequest],
) (*connect.Response[pb.DeleteWatermarkConfigResponse], error) {
	response, err := h.watermarkService.DeleteWatermarkConfig(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) ApplyWatermark(
	ctx context.Context,
	req *connect.Request[pb.ApplyWatermarkRequest],
) (*connect.Response[pb.ApplyWatermarkResponse], error) {
	response, err := h.watermarkService.ApplyWatermark(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

// Analytics handlers

func (h *DMSHandler) GetDocumentAccess(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentAccessRequest],
) (*connect.Response[pb.GetDocumentAccessResponse], error) {
	response, err := h.analyticsService.GetDocumentAccess(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetDocumentAnalytics(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentAnalyticsRequest],
) (*connect.Response[pb.GetDocumentAnalyticsResponse], error) {
	response, err := h.analyticsService.GetDocumentAnalytics(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetStorageUsage(
	ctx context.Context,
	req *connect.Request[pb.GetStorageUsageRequest],
) (*connect.Response[pb.GetStorageUsageResponse], error) {
	response, err := h.analyticsService.GetStorageUsage(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

// Document share handlers

func (h *DMSHandler) CreateDocumentShare(
	ctx context.Context,
	req *connect.Request[pb.CreateDocumentShareRequest],
) (*connect.Response[pb.CreateDocumentShareResponse], error) {
	response, err := h.shareService.CreateDocumentShare(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetDocumentShare(
	ctx context.Context,
	req *connect.Request[pb.GetDocumentShareRequest],
) (*connect.Response[pb.GetDocumentShareResponse], error) {
	response, err := h.shareService.GetDocumentShare(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) ListDocumentShares(
	ctx context.Context,
	req *connect.Request[pb.ListDocumentSharesRequest],
) (*connect.Response[pb.ListDocumentSharesResponse], error) {
	response, err := h.shareService.ListDocumentShares(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) GetUserShares(
	ctx context.Context,
	req *connect.Request[pb.GetUserSharesRequest],
) (*connect.Response[pb.GetUserSharesResponse], error) {
	response, err := h.shareService.GetUserShares(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *DMSHandler) DeleteDocumentShare(
	ctx context.Context,
	req *connect.Request[pb.DeleteDocumentShareRequest],
) (*connect.Response[pb.DeleteDocumentShareResponse], error) {
	response, err := h.shareService.DeleteDocumentShare(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}
