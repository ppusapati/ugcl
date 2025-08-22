package helper

import (
	"fmt"
	"strings"

	pb "p9e.in/ugcl/identity/api/v2/user"
	"p9e.in/ugcl/identity/models"
	query "p9e.in/ugcl/packages/api/v1/query"
	cmodels "p9e.in/ugcl/packages/models"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Add a constructor to create from proto request
func NewUserSearchCriteria(req *pb.ListUsersRequest) *models.UserSearchCriteria {
	criteria := &models.UserSearchCriteria{
		SearchCriteria: &cmodels.SearchCriteria{
			PageOffset: int32(req.PageOffset),
			PageSize:   int32(req.PageSize),
			Sort:       req.Sort,
		},
	}

	if req.Fields != nil {
		criteria.FieldMask = req.Fields
	}

	if req.Filter != nil {
		criteria.SearchTerm = &query.StringFilterOperation{Eq: wrapperspb.String(req.Filter.SearchTerm)}
		criteria.TenantIds = req.Filter.TenantIds
		criteria.RoleIds = req.Filter.RoleIds
		criteria.ActiveOnly = &query.BooleanFilterOperators{Eq: wrapperspb.Bool(req.Filter.ActiveOnly)}

		// // Handle AND conditions
		// if len(req.Filter.And) > 0 {
		// 	criteria.AndFilters = make([]cmodels.SearchCriteria, len(req.Filter.And))
		// 	for i, filter := range req.Filter.And {
		// 		subReq := &pb.ListUsersRequest{Filter: filter}
		// 		andCriteria := NewUserSearchCriteria(subReq)
		// 		criteria.AndFilters[i] = *andCriteria.SearchCriteria
		// 	}
		// }

		// // Handle OR conditions
		// if len(req.Filter.Or) > 0 {
		// 	criteria.OrFilters = make([]cmodels.SearchCriteria, len(req.Filter.Or))
		// 	for i, filter := range req.Filter.Or {
		// 		subReq := &pb.ListUsersRequest{Filter: filter}
		// 		orCriteria := NewUserSearchCriteria(subReq)
		// 		criteria.OrFilters[i] = *orCriteria.SearchCriteria
		// 	}
		// }
	}

	return criteria
}

// Helper function to filter user fields based on field mask
func FilterUserFields(user *pb.User, fields []string) *pb.User {
	if len(fields) == 0 {
		return user
	}

	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	// Create a new user with only the requested fields
	filtered := &pb.User{
		Id: user.Id, // Always include ID
	}

	// Only copy fields that are in the mask
	if fieldMap["username"] {
		filtered.Username = user.Username
	}
	if fieldMap["email"] {
		filtered.Email = user.Email
	}
	if fieldMap["fullname"] {
		filtered.Fullname = user.Fullname
	}
	if fieldMap["phone"] {
		filtered.Phone = user.Phone
	}
	if fieldMap["gender"] {
		filtered.Gender = user.Gender
	}
	if fieldMap["tenant_id"] {
		filtered.TenantId = user.TenantId
	}
	if fieldMap["roles"] {
		filtered.Roles = user.Roles
	}
	if fieldMap["avatar"] {
		filtered.Avatar = user.Avatar
	}
	if fieldMap["two_factor_enabled"] {
		filtered.TwoFactorEnabled = user.TwoFactorEnabled
	}
	if fieldMap["is_active"] {
		filtered.IsActive = user.IsActive
	}
	// Add other fields as needed

	return filtered
}

// BuildSelectFields constructs the SELECT part of the query based on FieldMask
func ExtractFieldMask(fieldMask []string, tableAlias string) string {
	if len(fieldMask) == 0 {
		return tableAlias + "*"
	}
	// Map proto fields to database columns
	fieldMapping := map[string]string{
		"id":        "id",
		"username":  "username",
		"phone":     "phone",
		"email":     "email",
		"fullname":  "fullname",
		"gender":    "gender",
		"is_active": "is_active",
		// Add more field mappings as needed
	}
	selectedFields := make([]string, 0, len(fieldMask))
	for _, field := range fieldMask {
		if dbField, ok := fieldMapping[field]; ok {
			selectedFields = append(selectedFields,
				fmt.Sprintf("%s.%s", tableAlias, dbField))
		}
	}
	if len(selectedFields) == 0 {
		return tableAlias + ".*"
	}
	return strings.Join(selectedFields, ", ")
}
