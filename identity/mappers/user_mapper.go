package mappers

import (
	pb "p9e.in/ugcl/identity/api/v2/user"
	sqlc "p9e.in/ugcl/identity/db/sqlc/generated"
	"p9e.in/ugcl/identity/models"

	// pmodels "packages/models"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func UserProtoToModel(pbUser *pb.User) *models.User {
	if pbUser == nil {
		return nil
	}
	strPtr := func(s *wrapperspb.StringValue) *string {
		if s == nil {
			return nil
		}
		val := s.GetValue()
		return &val
	}
	return &models.User{
		Uuid:     pbUser.GetUuid().GetValue(),
		Username: strPtr(pbUser.GetUsername()),
		Fullname: strPtr(pbUser.GetFullname()),
		Phone:    strPtr(pbUser.GetPhone()),
		Email:    strPtr(pbUser.GetEmail()),
		Password: pbUser.Password,
		Gender:   models.Gender(pbUser.Gender),
		// TenantId:         pbUser.TenantId,
		Roles:            pbUser.Roles,
		Avatar:           pbUser.Avatar,
		TwoFactorEnabled: pbUser.TwoFactorEnabled,
		TwoFactorSecret:  strPtr(pbUser.GetTwoFactorSecret()),
		IsActive:         pbUser.IsActive,
	}
}

func UserModelToProto(user *models.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Uuid:     &wrapperspb.StringValue{Value: user.Uuid},
		Username: &wrapperspb.StringValue{Value: *user.Username},
		Fullname: &wrapperspb.StringValue{Value: *user.Fullname},
		Phone:    &wrapperspb.StringValue{Value: *user.Phone},
		Email:    &wrapperspb.StringValue{Value: *user.Email},
		Password: user.Password,
		Gender:   pb.Gender(user.Gender),
		// TenantId:         user.TenantId,
		Roles:            user.Roles,
		TwoFactorEnabled: user.TwoFactorEnabled,
		TwoFactorSecret:  &wrapperspb.StringValue{Value: *user.TwoFactorSecret},
		EmailVerified:    user.EmailConfirmed,
		PhoneVerified:    user.PhoneConfirmed,
		IsActive:         user.IsActive,
	}
}

func UserModelToSQLC(user *models.User) sqlc.CreateUserParams {
	gender := int32(user.Gender)
	return sqlc.CreateUserParams{
		Uuid:             user.Uuid,
		Username:         user.Username,
		Fullname:         user.Fullname,
		Phone:            user.Phone,
		Email:            user.Email,
		Password:         &user.Password,
		Gender:           &gender,
		RolesCache:       user.Roles,
		TwoFactorEnabled: &user.TwoFactorEnabled,
		TwoFactorSecret:  user.TwoFactorSecret,
		EmailConfirmed:   &user.EmailConfirmed,
		PhoneConfirmed:   &user.PhoneConfirmed,
		IsActive:         user.IsActive,
	}
}

func UserSQLCToModel(row sqlc.User) *models.User {
	return &models.User{
		Uuid:             row.Uuid,
		Username:         row.Username,
		Fullname:         row.Fullname,
		Phone:            row.Phone,
		Email:            row.Email,
		Password:         *row.Password,
		Gender:           models.Gender(*row.Gender),
		Roles:            row.RolesCache,
		TwoFactorEnabled: *row.TwoFactorEnabled,
		TwoFactorSecret:  row.TwoFactorSecret,
		EmailConfirmed:   *row.EmailConfirmed,
		PhoneConfirmed:   *row.PhoneConfirmed,
		IsActive:         row.IsActive,
	}
}

func ProtoToUserIdentifier(pi *pb.UserIdentifier) *models.UserIdentifier {
	if pi == nil {
		return nil
	}

	var id int32
	var uuid string

	switch x := pi.Identifier.(type) {
	case *pb.UserIdentifier_Id:
		id = x.Id
	case *pb.UserIdentifier_Uuid:
		uuid = x.Uuid
	default:
		return nil
	}

	return &models.UserIdentifier{
		Id:   id,
		Uuid: uuid,
	}
}

func UserIdentifierToProto(mi *models.UserIdentifier) *pb.UserIdentifier {
	if mi == nil {
		return nil
	}

	switch {
	case mi.Id != 0:
		return &pb.UserIdentifier{
			Identifier: &pb.UserIdentifier_Id{Id: mi.Id},
		}
	case mi.Uuid != "":
		return &pb.UserIdentifier{
			Identifier: &pb.UserIdentifier_Uuid{Uuid: mi.Uuid},
		}
	default:
		return nil
	}
}

func UserProtoToSearchCriteria(req *pb.ListUsersRequest) *models.UserSearchCriteria {
	var username, phone, email, fullname *string

	if req.Filter != nil {
		if req.Filter.Username != nil && req.Filter.Username.GetValue() != "" {
			val := req.Filter.Username.GetValue()
			username = &val
		}
		if req.Filter.Phone != nil && req.Filter.Phone.GetValue() != "" {
			val := req.Filter.Phone.GetValue()
			phone = &val
		}
		if req.Filter.Email != nil && req.Filter.Email.GetValue() != "" {
			val := req.Filter.Email.GetValue()
			email = &val
		}
		if req.Filter.Fullname != nil && req.Filter.Fullname.GetValue() != "" {
			val := req.Filter.Fullname.GetValue()
			fullname = &val
		}
	}

	return &models.UserSearchCriteria{
		// SearchCriteria: &pmodels.SearchCriteria{
		// 	PageOffset: int32(req.PageOffset),
		// 	PageSize:   int32(req.GetPageSize()),
		// 	Sort:       req.GetSort(),
		// 	FieldMask:  req.GetFields(), // optional, handle nil if needed
		// },
		UsernameSearch: username,
		PhoneSearch:    phone,
		EmailSearch:    email,
		FullnameSearch: fullname,
	}
}
