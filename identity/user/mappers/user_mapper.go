package mappers

import (
	pb "p9e.in/ugcl/identity/user/api/v2/user"
	sqlc "p9e.in/ugcl/identity/user/db/sqlc/generated"
	"p9e.in/ugcl/identity/user/models"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	// Parse UUID from string
	var userUUID uuid.UUID
	if pbUser.GetUuid() != nil {
		parsedUUID, err := uuid.Parse(pbUser.GetUuid().GetValue())
		if err == nil {
			userUUID = parsedUUID
		}
	}

	return &models.User{
		Uuid:     userUUID,
		Username: strPtr(pbUser.GetUsername()),
		Fullname: strPtr(pbUser.GetFullname()),
		Phone:    strPtr(pbUser.GetPhone()),
		Email:    strPtr(pbUser.GetEmail()),
		Password: pbUser.Password,
		Gender:   models.Gender(pbUser.Gender),
		// TenantId:         pbUser.TenantId,
		TenantIds:        pbUser.TenantIds,
		TenantRoles:      convertTenantRolesToStringArray(pbUser.TenantRoles),
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

	pbUser := &pb.User{
		Id:               user.Id,
		Uuid:             &wrapperspb.StringValue{Value: user.Uuid.String()},
		Gender:           pb.Gender(user.Gender),
		TenantRoles:      convertStringArrayToTenantRoles(user.TenantRoles),
		TwoFactorEnabled: user.TwoFactorEnabled,
		EmailVerified:    user.EmailConfirmed,
		PhoneVerified:    user.PhoneConfirmed,
		IsActive:         user.IsActive,
		Avatar:           user.Avatar,
	}

	// Handle nullable fields
	if user.Username != nil {
		pbUser.Username = &wrapperspb.StringValue{Value: *user.Username}
	}
	if user.Fullname != nil {
		pbUser.Fullname = &wrapperspb.StringValue{Value: *user.Fullname}
	}
	if user.Phone != nil {
		pbUser.Phone = &wrapperspb.StringValue{Value: *user.Phone}
	}
	if user.Email != nil {
		pbUser.Email = &wrapperspb.StringValue{Value: *user.Email}
	}
	if user.TwoFactorSecret != nil {
		pbUser.TwoFactorSecret = &wrapperspb.StringValue{Value: *user.TwoFactorSecret}
	}

	return pbUser
}

func UserModelToSQLC(user *models.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		Uuid:               user.Uuid,
		Username:           user.Username,
		NormalizedUsername: user.NormalizedUsername,
		Fullname:           user.Fullname,
		Phone:              user.Phone,
		PhoneVerified:      &user.PhoneConfirmed,
		Email:              user.Email,
		NormalizedEmail:    user.NormalizedEmail,
		EmailVerified:      &user.EmailConfirmed,
		PasswordHash:       user.PasswordHash,
		Gender:             int32(user.Gender),
		RolesCache:         user.TenantRoles,
		Avatar:             user.Avatar,
		TwoFactorEnabled:   &user.TwoFactorEnabled,
		Salt:               user.Salt,
		TwoFactorSecret:    user.TwoFactorSecret,
		IsActive:           user.IsActive,
		Status:             nil, // Set to nil or appropriate default value
	}
}

func UserSQLCToModel(row sqlc.User) *models.User {
	// Handle interface{} gender field
	var gender models.Gender
	if row.Gender != nil {
		if g, ok := row.Gender.(int32); ok {
			gender = models.Gender(g)
		}
	}

	// Handle nullable fields
	var twoFactorEnabled bool
	if row.TwoFactorEnabled != nil {
		twoFactorEnabled = *row.TwoFactorEnabled
	}

	var emailConfirmed bool
	if row.EmailVerified != nil {
		emailConfirmed = *row.EmailVerified
	}

	var phoneConfirmed bool
	if row.PhoneVerified != nil {
		phoneConfirmed = *row.PhoneVerified
	}

	return &models.User{
		Uuid:               row.Uuid,
		Id:                 int32(row.ID),
		Username:           row.Username,
		NormalizedUsername: row.NormalizedUsername,
		Fullname:           row.Fullname,
		Phone:              row.Phone,
		Email:              row.Email,
		PasswordHash:       row.PasswordHash,
		Gender:             gender,
		TenantRoles:        row.RolesCache,
		Avatar:             row.Avatar,
		Salt:               row.Salt,
		TwoFactorEnabled:   twoFactorEnabled,
		TwoFactorSecret:    row.TwoFactorSecret,
		EmailConfirmed:     emailConfirmed,
		PhoneConfirmed:     phoneConfirmed,
		IsActive:           row.IsActive,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
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

func convertTenantRolesToStringArray(roles []*pb.UserTenantRole) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = role.String()
	}
	return result
}

func convertStringArrayToTenantRoles(roles []string) []*pb.UserTenantRole {
	result := make([]*pb.UserTenantRole, len(roles))
	for i, role := range roles {
		result[i] = &pb.UserTenantRole{
			TenantId:   "",                // This should be set by the caller or passed as parameter
			Roles:      []string{role},    // UserTenantRole.roles is repeated string
			IsActive:   true,              // Default to active
			AssignedAt: timestamppb.Now(), // Current timestamp
		}
	}
	return result
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
