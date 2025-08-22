package handlers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"p9e.in/ugcl/masters/pipeline/models"
	"p9e.in/ugcl/masters/pipeline/repository"
	"p9e.in/ugcl/masters/pipeline/services"

	pb "p9e.in/ugcl/masters/pipeline/api/v2/pipeline"

	"connectrpc.com/connect"
	"go.uber.org/fx"
)

type PipelineHandler struct {
	pipelineService  services.PipelineService
	hierarchyService services.HierarchyService
}

func NewPipelineHandler(
	pipelineService services.PipelineService,
	hierarchyService services.HierarchyService,
) *PipelineHandler {
	return &PipelineHandler{
		pipelineService:  pipelineService,
		hierarchyService: hierarchyService,
	}
}

// Fx Provider (optional: so you can import this as handlers.DairySiteHandlerModule)
var PipelineHandlerModule = fx.Provide(NewPipelineHandler)

func (h *PipelineHandler) GetPipelineSegments(
	ctx context.Context,
	req *connect.Request[pb.GetPipelineSegmentsRequest],
) (*connect.Response[pb.GetPipelineSegmentsResponse], error) {
	filter := repository.SegmentFilter{
		SiteID: req.Msg.SiteId,
		ZoneID: req.Msg.ZoneId,
		Label:  req.Msg.Label,
		Limit:  int(req.Msg.Limit),
		Offset: int(req.Msg.Offset),
	}

	segments, total, err := h.pipelineService.GetPipelineSegments(filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pipeline segments: %w", err))
	}

	pbSegments := make([]*pb.PipelineSegment, len(segments))
	for i, segment := range segments {
		pbSegment := &pb.PipelineSegment{
			Id:          segment.ID,
			SiteId:      segment.SiteID,
			ZoneId:      segment.ZoneID,
			Label:       segment.Label,
			StartNodeId: segment.StartNodeID,
			StopNodeId:  segment.StopNodeID,
			PipeId:      segment.PipeID,
			DiameterMm:  int32(segment.DiameterMM),
			LengthM:     segment.LengthM,
			Status:      string(segment.Status),
			CreatedAt:   timestamppb.New(segment.CreatedAt),
			UpdatedAt:   timestamppb.New(segment.UpdatedAt),
		}
		// Add related entities
		if segment.StartNode.ID != "" {
			pbSegment.StartNode = &pb.Node{
				Id:       segment.StartNode.ID,
				NodeName: segment.StartNode.NodeName,
				SiteId:   segment.StartNode.SiteID,
				NodeType: string(segment.StartNode.NodeType),
			}
		}

		if segment.StopNode.ID != "" {
			pbSegment.StopNode = &pb.Node{
				Id:       segment.StopNode.ID,
				NodeName: segment.StopNode.NodeName,
				SiteId:   segment.StopNode.SiteID,
				NodeType: string(segment.StopNode.NodeType),
			}
		}

		if segment.Pipe.ID != "" {
			pbSegment.Pipe = &pb.Pipe{
				Id:             segment.Pipe.ID,
				PipeName:       segment.Pipe.PipeName,
				PipeType:       string(segment.Pipe.PipeType),
				PressureRating: segment.Pipe.PressureRating,
				Description:    segment.Pipe.Description,
			}
		}

		pbSegments[i] = pbSegment
	}

	response := &pb.GetPipelineSegmentsResponse{
		Segments:   pbSegments,
		TotalCount: int32(total),
	}

	return connect.NewResponse(response), nil
}

func (h *PipelineHandler) CreatePipelineSegment(
	ctx context.Context,
	req *connect.Request[pb.CreatePipelineSegmentRequest],
) (*connect.Response[pb.CreatePipelineSegmentResponse], error) {
	segment := &models.PipelineSegment{
		SiteID:      req.Msg.SiteId,
		ZoneID:      req.Msg.ZoneId,
		Label:       req.Msg.Label,
		StartNodeID: req.Msg.StartNodeId,
		StopNodeID:  req.Msg.StopNodeId,
		PipeID:      req.Msg.PipeId,
		DiameterMM:  int(req.Msg.DiameterMm),
		LengthM:     req.Msg.LengthM,
		Status:      models.SegmentStatus(req.Msg.Status),
	}

	if req.Msg.InstallationDate != "" {
		installDate, err := time.Parse("2006-01-02", req.Msg.InstallationDate)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid installation date format: %w", err))
		}
		segment.InstallationDate = &installDate
	}

	err := h.pipelineService.CreatePipelineSegment(segment)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create pipeline segment: %w", err))
	}

	// Fetch the created segment with all relationships
	createdSegment, err := h.pipelineService.GetPipelineSegmentByID(segment.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch created segment: %w", err))
	}

	pbSegment := &pb.PipelineSegment{
		Id:          createdSegment.ID,
		SiteId:      createdSegment.SiteID,
		ZoneId:      createdSegment.ZoneID,
		Label:       createdSegment.Label,
		StartNodeId: createdSegment.StartNodeID,
		StopNodeId:  createdSegment.StopNodeID,
		PipeId:      createdSegment.PipeID,
		DiameterMm:  int32(createdSegment.DiameterMM),
		LengthM:     createdSegment.LengthM,
		Status:      string(createdSegment.Status),
		CreatedAt:   timestamppb.New(createdSegment.CreatedAt),
		UpdatedAt:   timestamppb.New(createdSegment.UpdatedAt),
	}
	response := &pb.CreatePipelineSegmentResponse{
		Segment: pbSegment,
	}

	return connect.NewResponse(response), nil
}

func (h *PipelineHandler) GetSites(
	ctx context.Context,
	req *connect.Request[pb.GetSitesRequest],
) (*connect.Response[pb.GetSitesResponse], error) {
	sites, err := h.hierarchyService.GetSites()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get sites: %w", err))
	}

	pbSites := make([]*pb.Site, len(sites))
	for i, site := range sites {
		pbSites[i] = &pb.Site{
			Id:          site.ID,
			SiteName:    site.SiteName,
			SiteCode:    site.SiteCode,
			Description: site.Description,
			CreatedAt:   timestamppb.New(site.CreatedAt),
			UpdatedAt:   timestamppb.New(site.UpdatedAt),
			ZoneCount:   int32(site.ZoneCount),
		}
	}

	response := &pb.GetSitesResponse{
		Sites: pbSites,
	}

	return connect.NewResponse(response), nil
}

// ✅ IMPLEMENTED: GetZonesBySite
func (h *PipelineHandler) GetZonesBySite(
	ctx context.Context,
	req *connect.Request[pb.GetZonesBySiteRequest],
) (*connect.Response[pb.GetZonesBySiteResponse], error) {
	var zones []repository.ZoneWithCounts
	var err error

	// Support both ID and UUID
	if req.Msg.SiteId == "" {
		zones, err = h.hierarchyService.GetZonesBySite(req.Msg.SiteId)
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("either site_id or site_uuid must be provided"))
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get zones: %w", err))
	}

	pbZones := make([]*pb.Zone, len(zones))
	for i, zone := range zones {
		pbZones[i] = &pb.Zone{
			Id:           zone.ID,
			SiteId:       zone.SiteID,
			ZoneName:     zone.ZoneName,
			ZoneCode:     zone.ZoneCode,
			Description:  zone.Description,
			CreatedAt:    timestamppb.New(zone.CreatedAt),
			UpdatedAt:    timestamppb.New(zone.UpdatedAt),
			LabelCount:   int32(zone.LabelCount),
			SegmentCount: int32(zone.SegmentCount),
		}
	}

	response := &pb.GetZonesBySiteResponse{
		Zones: pbZones,
	}

	return connect.NewResponse(response), nil
}

// ✅ IMPLEMENTED: GetLabelsByZone
func (h *PipelineHandler) GetLabelsByZone(
	ctx context.Context,
	req *connect.Request[pb.GetLabelsByZoneRequest],
) (*connect.Response[pb.GetLabelsByZoneResponse], error) {
	var labels []repository.LabelInfo
	var err error

	// Support both ID and UUID
	if req.Msg.ZoneId == "" {
		labels, err = h.hierarchyService.GetLabelsByZone(req.Msg.ZoneId)
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("either zone_id or zone_uuid must be provided"))
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get labels: %w", err))
	}

	pbLabels := make([]*pb.LabelInfo, len(labels))
	for i, label := range labels {
		pbLabels[i] = &pb.LabelInfo{
			Label:        label.Label,
			SegmentCount: int32(label.SegmentCount),
			TotalLength:  label.TotalLength,
			PipeUsed:     label.PipeUsed,
		}
	}

	response := &pb.GetLabelsByZoneResponse{
		Labels: pbLabels,
	}

	return connect.NewResponse(response), nil
}

// ✅ IMPLEMENTED: GetNodesByZone
func (h *PipelineHandler) GetNodesByZone(
	ctx context.Context,
	req *connect.Request[pb.GetNodesByZoneRequest],
) (*connect.Response[pb.GetNodesByZoneResponse], error) {
	var nodes []models.Node
	var err error

	// Support both ID and UUID
	if req.Msg.ZoneId == "" {
		nodes, err = h.hierarchyService.GetNodesByZone(req.Msg.ZoneId)
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("either zone_id or zone_uuid must be provided"))
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get nodes: %w", err))
	}

	pbNodes := make([]*pb.Node, len(nodes))
	for i, node := range nodes {
		pbNode := &pb.Node{
			Id:        node.ID,
			NodeName:  node.NodeName,
			SiteId:    node.SiteID,
			NodeType:  string(node.NodeType),
			CreatedAt: timestamppb.New(node.CreatedAt),
		}

		if node.ZoneID != nil {
			pbNode.ZoneId = *node.ZoneID
		}
		if node.CoordinatesX != nil {
			pbNode.CoordinatesX = *node.CoordinatesX
		}
		if node.CoordinatesY != nil {
			pbNode.CoordinatesY = *node.CoordinatesY
		}
		if node.Elevation != nil {
			pbNode.Elevation = *node.Elevation
		}

		pbNodes[i] = pbNode
	}

	response := &pb.GetNodesByZoneResponse{
		Nodes: pbNodes,
	}

	return connect.NewResponse(response), nil
}

// ✅ IMPLEMENTED: GetNodesByLabel
func (h *PipelineHandler) GetNodesByLabel(
	ctx context.Context,
	req *connect.Request[pb.GetNodesByLabelRequest],
) (*connect.Response[pb.GetNodesByLabelResponse], error) {
	if req.Msg.Label == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("label must be provided"))
	}

	nodes, err := h.hierarchyService.GetNodesByLabel(req.Msg.Label)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get nodes: %w", err))
	}

	pbNodes := make([]*pb.Node, len(nodes))
	for i, node := range nodes {
		pbNode := &pb.Node{
			Id:        node.ID,
			NodeName:  node.NodeName,
			SiteId:    node.SiteID,
			NodeType:  string(node.NodeType),
			CreatedAt: timestamppb.New(node.CreatedAt),
		}

		if node.ZoneID != nil {
			pbNode.ZoneId = *node.ZoneID
		}
		if node.CoordinatesX != nil {
			pbNode.CoordinatesX = *node.CoordinatesX
		}
		if node.CoordinatesY != nil {
			pbNode.CoordinatesY = *node.CoordinatesY
		}
		if node.Elevation != nil {
			pbNode.Elevation = *node.Elevation
		}

		pbNodes[i] = pbNode
	}

	response := &pb.GetNodesByLabelResponse{
		Nodes: pbNodes,
	}

	return connect.NewResponse(response), nil
}

func (h *PipelineHandler) GetPipelineView(
	ctx context.Context,
	req *connect.Request[pb.GetPipelineViewRequest],
) (*connect.Response[pb.GetPipelineViewResponse], error) {
	segments, err := h.pipelineService.GetPipelineView(req.Msg.SiteName, req.Msg.ZoneName, req.Msg.Label)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pipeline view: %w", err))
	}

	pbSegments := make([]*pb.PipelineSegmentView, len(segments))
	for i, segment := range segments {
		pbSegments[i] = &pb.PipelineSegmentView{
			SiteName:  segment.SiteName,
			ZoneName:  segment.ZoneName,
			Label:     segment.Label,
			StartNode: segment.StartNode,
			StopNode:  segment.StopNode,
			Diameter:  int32(segment.Diameter),
			Length:    segment.Length,
			PipeName:  segment.PipeName,
			CreatedAt: timestamppb.New(segment.CreatedAt),
			UpdatedAt: timestamppb.New(segment.UpdatedAt),
		}
	}

	response := &pb.GetPipelineViewResponse{
		Segments: pbSegments,
	}

	return connect.NewResponse(response), nil
}
