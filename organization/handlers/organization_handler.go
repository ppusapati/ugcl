package handlers

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "p9e.in/ugcl/organization/api/v1/organization"
	"p9e.in/ugcl/organization/api/v1/organization/organizationconnect"
	"p9e.in/ugcl/organization/services"
)

type OrganizationHandler struct {
	divisionService   services.IDivisionService
	branchService     services.IBranchService
	departmentService services.IDepartmentService
}

// NewOrganizationHandler creates a new organization Connect handler
func NewOrganizationHandler(
	divisionService services.IDivisionService,
	branchService services.IBranchService,
	departmentService services.IDepartmentService,
) organizationconnect.OrganizationServiceHandler {
	return &OrganizationHandler{
		divisionService:   divisionService,
		branchService:     branchService,
		departmentService: departmentService,
	}
}

// Division Operations

func (h *OrganizationHandler) CreateDivision(
	ctx context.Context,
	req *connect.Request[pb.CreateDivisionRequest],
) (*connect.Response[pb.Division], error) {
	division, err := h.divisionService.CreateDivision(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(division), nil
}

func (h *OrganizationHandler) UpdateDivision(
	ctx context.Context,
	req *connect.Request[pb.UpdateDivisionRequest],
) (*connect.Response[pb.Division], error) {
	division, err := h.divisionService.UpdateDivision(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(division), nil
}

func (h *OrganizationHandler) GetDivision(
	ctx context.Context,
	req *connect.Request[pb.GetDivisionRequest],
) (*connect.Response[pb.Division], error) {
	division, err := h.divisionService.GetDivision(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(division), nil
}

func (h *OrganizationHandler) GetDivisionByCode(
	ctx context.Context,
	req *connect.Request[pb.GetDivisionByCodeRequest],
) (*connect.Response[pb.Division], error) {
	division, err := h.divisionService.GetDivisionByCode(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(division), nil
}

func (h *OrganizationHandler) ListDivisions(
	ctx context.Context,
	req *connect.Request[pb.ListDivisionsRequest],
) (*connect.Response[pb.ListDivisionsResponse], error) {
	response, err := h.divisionService.ListDivisions(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListDivisionSummaries(
	ctx context.Context,
	req *connect.Request[pb.ListDivisionSummariesRequest],
) (*connect.Response[pb.ListDivisionSummariesResponse], error) {
	response, err := h.divisionService.ListDivisionSummaries(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) DeleteDivision(
	ctx context.Context,
	req *connect.Request[pb.DeleteDivisionRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.divisionService.DeleteDivision(ctx, req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Branch Operations

func (h *OrganizationHandler) CreateBranch(
	ctx context.Context,
	req *connect.Request[pb.CreateBranchRequest],
) (*connect.Response[pb.Branch], error) {
	branch, err := h.branchService.CreateBranch(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(branch), nil
}

func (h *OrganizationHandler) UpdateBranch(
	ctx context.Context,
	req *connect.Request[pb.UpdateBranchRequest],
) (*connect.Response[pb.Branch], error) {
	branch, err := h.branchService.UpdateBranch(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(branch), nil
}

func (h *OrganizationHandler) GetBranch(
	ctx context.Context,
	req *connect.Request[pb.GetBranchRequest],
) (*connect.Response[pb.Branch], error) {
	branch, err := h.branchService.GetBranch(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(branch), nil
}

func (h *OrganizationHandler) GetBranchByCode(
	ctx context.Context,
	req *connect.Request[pb.GetBranchByCodeRequest],
) (*connect.Response[pb.Branch], error) {
	branch, err := h.branchService.GetBranchByCode(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(branch), nil
}

func (h *OrganizationHandler) ListBranches(
	ctx context.Context,
	req *connect.Request[pb.ListBranchesRequest],
) (*connect.Response[pb.ListBranchesResponse], error) {
	response, err := h.branchService.ListBranches(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListBranchesByDivision(
	ctx context.Context,
	req *connect.Request[pb.ListBranchesByDivisionRequest],
) (*connect.Response[pb.ListBranchesByDivisionResponse], error) {
	response, err := h.branchService.ListBranchesByDivision(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListBranchesByCity(
	ctx context.Context,
	req *connect.Request[pb.ListBranchesByCityRequest],
) (*connect.Response[pb.ListBranchesResponse], error) {
	response, err := h.branchService.ListBranchesByCity(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListBranchesByState(
	ctx context.Context,
	req *connect.Request[pb.ListBranchesByStateRequest],
) (*connect.Response[pb.ListBranchesResponse], error) {
	response, err := h.branchService.ListBranchesByState(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListBranchSummaries(
	ctx context.Context,
	req *connect.Request[pb.ListBranchSummariesRequest],
) (*connect.Response[pb.ListBranchSummariesResponse], error) {
	response, err := h.branchService.ListBranchSummaries(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) DeleteBranch(
	ctx context.Context,
	req *connect.Request[pb.DeleteBranchRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.branchService.DeleteBranch(ctx, req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Department Operations

func (h *OrganizationHandler) CreateDepartment(
	ctx context.Context,
	req *connect.Request[pb.CreateDepartmentRequest],
) (*connect.Response[pb.Department], error) {
	department, err := h.departmentService.CreateDepartment(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(department), nil
}

func (h *OrganizationHandler) UpdateDepartment(
	ctx context.Context,
	req *connect.Request[pb.UpdateDepartmentRequest],
) (*connect.Response[pb.Department], error) {
	department, err := h.departmentService.UpdateDepartment(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(department), nil
}

func (h *OrganizationHandler) GetDepartment(
	ctx context.Context,
	req *connect.Request[pb.GetDepartmentRequest],
) (*connect.Response[pb.Department], error) {
	department, err := h.departmentService.GetDepartment(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(department), nil
}

func (h *OrganizationHandler) GetDepartmentByCode(
	ctx context.Context,
	req *connect.Request[pb.GetDepartmentByCodeRequest],
) (*connect.Response[pb.Department], error) {
	department, err := h.departmentService.GetDepartmentByCode(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(department), nil
}

func (h *OrganizationHandler) ListDepartments(
	ctx context.Context,
	req *connect.Request[pb.ListDepartmentsRequest],
) (*connect.Response[pb.ListDepartmentsResponse], error) {
	response, err := h.departmentService.ListDepartments(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListDepartmentsByDivision(
	ctx context.Context,
	req *connect.Request[pb.ListDepartmentsByDivisionRequest],
) (*connect.Response[pb.ListDepartmentsResponse], error) {
	response, err := h.departmentService.ListDepartmentsByDivision(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListBusinessLevelDepartments(
	ctx context.Context,
	req *connect.Request[pb.ListBusinessLevelDepartmentsRequest],
) (*connect.Response[pb.ListDepartmentsResponse], error) {
	response, err := h.departmentService.ListBusinessLevelDepartments(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) ListSubDepartments(
	ctx context.Context,
	req *connect.Request[pb.ListSubDepartmentsRequest],
) (*connect.Response[pb.ListDepartmentsResponse], error) {
	response, err := h.departmentService.ListSubDepartments(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) GetDepartmentHierarchy(
	ctx context.Context,
	req *connect.Request[pb.GetDepartmentHierarchyRequest],
) (*connect.Response[pb.GetDepartmentHierarchyResponse], error) {
	response, err := h.departmentService.GetDepartmentHierarchy(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (h *OrganizationHandler) DeleteDepartment(
	ctx context.Context,
	req *connect.Request[pb.DeleteDepartmentRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.departmentService.DeleteDepartment(ctx, req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}
