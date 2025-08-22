// handlers/user_handler.go
package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"p9e.in/ugcl/identity/api/v2/user"
	"p9e.in/ugcl/identity/mappers"
	"p9e.in/ugcl/identity/services"
)

type UserHandler struct {
	// userpb.UnimplementedUserServiceServer
	Service services.IUserService
}

func NewUserHandler(service services.IUserService) *UserHandler {
	return &UserHandler{Service: service}
}

// AssignRoleToUser implements userconnect.UserServiceHandler.
func (u *UserHandler) AssignRoleToUser(ctx context.Context, req *connect.Request[user.AssignRoleRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, u.Service.AssignUserToRole(ctx, req.Msg.UserId, req.Msg.RoleId)
}

// CreateUser implements userconnect.UserServiceHandler.
func (u *UserHandler) CreateUser(ctx context.Context, req *connect.Request[user.CreateUserRequest]) (*connect.Response[user.UserIdentifier], error) {
	res, err := u.Service.RegisterUser(ctx, mappers.UserProtoToModel(req.Msg.User))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&user.UserIdentifier{
		Identifier: &user.UserIdentifier_Uuid{
			Uuid: mappers.UserModelToProto(res).Uuid.String(),
		},
	}), nil
}

// DeleteUser implements userconnect.UserServiceHandler.
func (u *UserHandler) DeleteUser(ctx context.Context, req *connect.Request[user.DeleteUserRequest]) (*connect.Response[user.DeleteUserResponse], error) {
	err := u.Service.DeleteUser(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&user.DeleteUserResponse{}), nil
}

// GetUser implements userconnect.UserServiceHandler.
func (u *UserHandler) GetUser(ctx context.Context, req *connect.Request[user.UserIdentifier]) (*connect.Response[user.User], error) {
	res, err := u.Service.GetByIdentifier(ctx, mappers.ProtoToUserIdentifier(req.Msg))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(mappers.UserModelToProto(res)), nil
}

// ListUsers implements userconnect.UserServiceHandler.
func (u *UserHandler) ListUsers(ctx context.Context, req *connect.Request[user.ListUsersRequest]) (*connect.Response[user.ListUsersResponse], error) {
	fmt.Println(req.Msg)
	res, err := u.Service.ListUsers(ctx, mappers.UserProtoToSearchCriteria(req.Msg))
	if err != nil {
		return nil, err
	}
	usersProto := make([]*user.User, len(res))
	for i, u := range res {
		usersProto[i] = mappers.UserModelToProto(u)
	}
	return connect.NewResponse(&user.ListUsersResponse{
		Users: usersProto,
	}), nil
}

// RevokeRoleFromUser implements userconnect.UserServiceHandler.
func (u *UserHandler) RevokeRoleFromUser(ctx context.Context, req *connect.Request[user.AssignRoleRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, u.Service.RevokeUserFromRole(ctx, req.Msg.UserId, req.Msg.RoleId)
}

// UpdateUser implements userconnect.UserServiceHandler.
func (u *UserHandler) UpdateUser(ctx context.Context, req *connect.Request[user.UpdateUserRequest]) (*connect.Response[user.UserIdentifier], error) {
	err := u.Service.UpdateUser(ctx, mappers.UserProtoToModel(req.Msg.User))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&user.UserIdentifier{
		Identifier: &user.UserIdentifier_Uuid{
			Uuid: req.Msg.User.Uuid.String(),
		},
	}), nil
}

// func (h *UserHandler) GetUserRoles(ctx context.Context, req *userpb.UserIdentifier) (*userpb.GetUserRoleResponse, error) {
// 	return h.Service.GetUserRoles(ctx, mappers.ProtoToUserIdentifier(req).Uuid)
// }
