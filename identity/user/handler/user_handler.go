// handlers/user_handler.go
package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"p9e.in/ugcl/identity/user/api/v2/user"
	"p9e.in/ugcl/identity/user/mappers"
	"p9e.in/ugcl/identity/user/services"
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

// Password Management
func (u *UserHandler) ChangePassword(ctx context.Context, req *connect.Request[user.ChangePasswordRequest]) (*connect.Response[emptypb.Empty], error) {
	err := u.Service.ChangePassword(ctx, req.Msg.UserId, req.Msg.CurrentPassword, req.Msg.NewPassword, req.Msg.ConfirmPassword)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) InitiatePasswordReset(ctx context.Context, req *connect.Request[user.InitiatePasswordResetRequest]) (*connect.Response[user.InitiatePasswordResetResponse], error) {
	// TODO: Implement password reset initiation
	return connect.NewResponse(&user.InitiatePasswordResetResponse{
		ResetToken:     "dummy_reset_token",
		Method:         user.PasswordResetMethod_PASSWORD_RESET_METHOD_EMAIL,
		MaskedContact:  "j***@example.com",
	}), nil
}

func (u *UserHandler) ResetPassword(ctx context.Context, req *connect.Request[user.ResetPasswordRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement password reset
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) ValidatePasswordStrength(ctx context.Context, req *connect.Request[user.ValidatePasswordStrengthRequest]) (*connect.Response[user.ValidatePasswordStrengthResponse], error) {
	// TODO: Implement password strength validation
	return connect.NewResponse(&user.ValidatePasswordStrengthResponse{
		IsValid:         true,
		StrengthScore:   85,
		StrengthLevel:   user.PasswordStrength_PASSWORD_STRENGTH_STRONG,
		RequirementsMet: []string{"length", "uppercase", "lowercase", "numbers"},
	}), nil
}

// Email/Phone Verification
func (u *UserHandler) SendVerificationEmail(ctx context.Context, req *connect.Request[user.SendVerificationRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement email verification sending
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) VerifyEmail(ctx context.Context, req *connect.Request[user.VerifyEmailRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement email verification
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) SendVerificationSMS(ctx context.Context, req *connect.Request[user.SendVerificationRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement SMS verification sending
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) VerifyPhone(ctx context.Context, req *connect.Request[user.VerifyPhoneRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement phone verification
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// User Profile Management
func (u *UserHandler) UpdateProfile(ctx context.Context, req *connect.Request[user.UpdateProfileRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement profile update
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (u *UserHandler) UploadAvatar(ctx context.Context, req *connect.Request[user.UploadAvatarRequest]) (*connect.Response[user.UploadAvatarResponse], error) {
	// TODO: Implement avatar upload
	return connect.NewResponse(&user.UploadAvatarResponse{
		AvatarUrl: "https://example.com/avatars/user_avatar.jpg",
		AvatarId:  "avatar_123",
	}), nil
}

func (u *UserHandler) DeleteAvatar(ctx context.Context, req *connect.Request[user.UserIdentifier]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement avatar deletion
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// User Preferences
func (u *UserHandler) GetUserPreferences(ctx context.Context, req *connect.Request[user.UserIdentifier]) (*connect.Response[user.UserPreferencesResponse], error) {
	// TODO: Implement get user preferences
	return connect.NewResponse(&user.UserPreferencesResponse{
		Preferences: &user.UserPreferences{
			Language:           "en",
			Timezone:           "UTC",
			DateFormat:         "YYYY-MM-DD",
			TimeFormat:         "24h",
			EmailNotifications: true,
			SmsNotifications:   false,
			PushNotifications:  true,
		},
	}), nil
}

func (u *UserHandler) UpdateUserPreferences(ctx context.Context, req *connect.Request[user.UpdateUserPreferencesRequest]) (*connect.Response[emptypb.Empty], error) {
	// TODO: Implement update user preferences
	return connect.NewResponse(&emptypb.Empty{}), nil
}
