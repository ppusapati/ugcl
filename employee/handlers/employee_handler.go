package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "p9e.in/ugcl/employee/api/v2/employee"
	"p9e.in/ugcl/employee/mappers"
	"p9e.in/ugcl/employee/services"
)

type EmployeeHandler struct {
	srvc services.IEmployeeService
}

func NewEmployeeHandler(srvc services.IEmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		srvc: srvc,
	}
}

// CreateEmployee implements EmployeeService.CreateEmployee
func (h *EmployeeHandler) CreateEmployee(
	ctx context.Context,
	req *connect.Request[pb.CreateEmployeeRequest],
) (*connect.Response[pb.Employee], error) {

	// Validate or generate UUID
	if req.Msg.Employee.Id == "" {
		req.Msg.Employee.Id = uuid.New().String()
	}

	// Map from proto to DB struct
	dbEmployee, err := mappers.ProtoToDBEmployee(req.Msg.Employee)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call service to create user and employee
	createdEmployee, err := h.srvc.Create(ctx, dbEmployee, req.Msg.User)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back to proto
	employeeProto, err := mappers.DBToProtoEmployee(createdEmployee)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(employeeProto), nil
}

// UpdateEmployee implements EmployeeService.UpdateEmployee
func (h *EmployeeHandler) UpdateEmployee(
	ctx context.Context,
	req *connect.Request[pb.UpdateEmployeeRequest],
) (*connect.Response[pb.Employee], error) {

	// Fetch existing employee
	existing, err := h.srvc.GetByID(ctx, req.Msg.Employee.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Apply field mask
	updated := mappers.ApplyFieldMask(existing, req.Msg.Employee, req.Msg.UpdateMask)

	// Persist
	updatedEmployee, err := h.srvc.Update(ctx, updated)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map back
	protoResp, err := mappers.DBToProtoEmployee(updatedEmployee)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoResp), nil
}

// GetEmployee implements EmployeeService.GetEmployee
func (h *EmployeeHandler) GetEmployee(
	ctx context.Context,
	req *connect.Request[pb.EmployeeIdentifier],
) (*connect.Response[pb.Employee], error) {

	var dbEmployee *pb.Employee
	var err error

	// Handle different identifier types
	switch id := req.Msg.Identifier.(type) {
	case *pb.EmployeeIdentifier_Id:
		emp, e := h.srvc.GetByID(ctx, id.Id)
		if e != nil {
			return nil, connect.NewError(connect.CodeNotFound, e)
		}
		dbEmployee, err = mappers.DBToProtoEmployee(emp)
	case *pb.EmployeeIdentifier_EmployeeCode:
		emp, e := h.srvc.GetByCode(ctx, id.EmployeeCode)
		if e != nil {
			return nil, connect.NewError(connect.CodeNotFound, e)
		}
		dbEmployee, err = mappers.DBToProtoEmployee(emp)
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(dbEmployee), nil
}

// ListEmployees implements EmployeeService.ListEmployees
func (h *EmployeeHandler) ListEmployees(
	ctx context.Context,
	req *connect.Request[pb.ListEmployeesRequest],
) (*connect.Response[pb.ListEmployeesResponse], error) {

	dbList, total, err := h.srvc.List(ctx, req.Msg.PageSize, req.Msg.PageOffset, req.Msg.Filter, req.Msg.Sort)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Employee, len(dbList))
	for i, e := range dbList {
		pe, err := mappers.DBToProtoEmployee(&e)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pe
	}

	return connect.NewResponse(&pb.ListEmployeesResponse{
		Employees:  protoList,
		TotalCount: int32(total),
	}), nil
}

// DeleteEmployee implements EmployeeService.DeleteEmployee
func (h *EmployeeHandler) DeleteEmployee(
	ctx context.Context,
	req *connect.Request[pb.EmployeeIdentifier],
) (*connect.Response[emptypb.Empty], error) {

	// Handle different identifier types
	switch id := req.Msg.Identifier.(type) {
	case *pb.EmployeeIdentifier_Id:
		if err := h.srvc.Delete(ctx, id.Id); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	case *pb.EmployeeIdentifier_EmployeeCode:
		// Get employee by code first to get ID
		emp, err := h.srvc.GetByCode(ctx, id.EmployeeCode)
		if err != nil {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		employeeID := mappers.PgUuidToStr(emp.Uuid)
		if err := h.srvc.Delete(ctx, employeeID); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, nil)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetEmployeeByUserId implements EmployeeService.GetEmployeeByUserId
func (h *EmployeeHandler) GetEmployeeByUserId(
	ctx context.Context,
	req *connect.Request[pb.UserIdentifier],
) (*connect.Response[pb.Employee], error) {

	dbEmployee, err := h.srvc.GetByUserID(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	protoEmployee, err := mappers.DBToProtoEmployee(dbEmployee)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(protoEmployee), nil
}

// GetEmployeesByDepartment implements EmployeeService.GetEmployeesByDepartment
func (h *EmployeeHandler) GetEmployeesByDepartment(
	ctx context.Context,
	req *connect.Request[pb.DepartmentIdentifier],
) (*connect.Response[pb.ListEmployeesResponse], error) {

	dbList, err := h.srvc.GetByDepartment(ctx, req.Msg.DepartmentId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Employee, len(dbList))
	for i, e := range dbList {
		pe, err := mappers.DBToProtoEmployee(&e)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pe
	}

	return connect.NewResponse(&pb.ListEmployeesResponse{
		Employees:  protoList,
		TotalCount: int32(len(protoList)),
	}), nil
}

// GetEmployeesByManager implements EmployeeService.GetEmployeesByManager
func (h *EmployeeHandler) GetEmployeesByManager(
	ctx context.Context,
	req *connect.Request[pb.ManagerIdentifier],
) (*connect.Response[pb.ListEmployeesResponse], error) {

	dbList, err := h.srvc.GetByManager(ctx, req.Msg.ManagerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoList := make([]*pb.Employee, len(dbList))
	for i, e := range dbList {
		pe, err := mappers.DBToProtoEmployee(&e)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoList[i] = pe
	}

	return connect.NewResponse(&pb.ListEmployeesResponse{
		Employees:  protoList,
		TotalCount: int32(len(protoList)),
	}), nil
}
