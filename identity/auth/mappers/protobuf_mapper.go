package mappers

// TODO: Fix all protobuf mapper functions to match current protobuf schema
// The current protobuf schema has type mismatches and missing enum values
// This file needs to be updated when the protobuf schema is finalized

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	authpb "p9e.in/ugcl/identity/auth/api/v2/auth"
	"p9e.in/ugcl/identity/auth/models"
	"p9e.in/ugcl/identity/auth/services/interfaces"
)

type ProtobufMapper struct{}

func NewProtobufMapper() *ProtobufMapper {
	return &ProtobufMapper{}
}

// Stub implementations for protobuf mapping
func (m *ProtobufMapper) ExtractIdentifierFromLoginRequest(req *authpb.LoginRequest) (string, string) {
	switch identifier := req.Identifier.(type) {
	case *authpb.LoginRequest_Email:
		return "email", identifier.Email
	case *authpb.LoginRequest_Username:
		return "username", identifier.Username
	case *authpb.LoginRequest_Phone:
		return "phone", identifier.Phone
	default:
		return "", ""
	}
}

func (m *ProtobufMapper) ExtractIdentifierFromOTPRequest(req *authpb.LoginOTPRequest) (string, string) {
	switch identifier := req.Identifier.(type) {
	case *authpb.LoginOTPRequest_Email:
		return "email", identifier.Email
	case *authpb.LoginOTPRequest_Username:
		return "username", identifier.Username
	case *authpb.LoginOTPRequest_Phone:
		return "phone", identifier.Phone
	default:
		return "", ""
	}
}

func (m *ProtobufMapper) BuildLoginResponse(resp *interfaces.LoginResponse) *authpb.LoginResponse {
	response := &authpb.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		// ExpiresIn would need to be calculated from ExpiresAt if needed
	}

	if resp.User != nil {
		response.User = &authpb.UserSession{
			UserId:   resp.User.UserID,
			Username: resp.User.Username,
			Email:    resp.User.Email,
		}
		// TenantID would need to be extracted from UserTenants or Session if needed
	}

	return response
}

func (m *ProtobufMapper) ToProtobufSession(session *models.Session) *authpb.Session {
	pbSession := &authpb.Session{
		SessionId: session.SessionID,
		UserId:    session.UserID.String(),
		IsActive:  session.IsActive,
		ExpiresAt: timestamppb.New(session.ExpiresAt),
		CreatedAt: timestamppb.New(session.CreatedAt),
	}

	if session.TenantID != nil {
		pbSession.TenantId = session.TenantID
	}
	if session.IPAddress != nil {
		pbSession.IpAddress = session.IPAddress
	}
	if session.LastAccessedAt != nil {
		pbSession.LastAccessed = timestamppb.New(*session.LastAccessedAt)
	}

	return pbSession
}