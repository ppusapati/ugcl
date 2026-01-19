package dairysitemapper

import (
	"github.com/jackc/pgx/v5/pgtype"
	"p9e.in/ugcl/projects/api/v2/dairy_site"
	db "p9e.in/ugcl/projects/db/generated"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToCreateParams maps from proto to sqlc CreateDairySiteParams
func ProtoToCreateParams(proto *dairy_site.DairySite) db.CreateDairySiteParams {
	var submittedAt pgtype.Timestamptz
	if proto.SubmittedAt != nil {
		submittedAt = pgtype.Timestamptz{Time: proto.SubmittedAt.AsTime(), Valid: true}
	}
	return db.CreateDairySiteParams{
		NameOfSite:  proto.NameOfSite,
		TodaysWork:  &proto.TodaysWork,
		EmployeeID:  proto.EmployeeId,
		Latitude:    &proto.Latitude,
		Longitude:   &proto.Longitude,
		SubmittedAt: submittedAt,
	}
}

// DBToProto maps sqlc DairySite to proto DairySite
func DBToProto(dbSite db.DairySite) *dairy_site.DairySite {
	return &dairy_site.DairySite{
		Id:          dbSite.ID,
		NameOfSite:  dbSite.NameOfSite,
		TodaysWork:  *dbSite.TodaysWork,
		EmployeeId:  dbSite.EmployeeID,
		Latitude:    *dbSite.Latitude,
		Longitude:   *dbSite.Longitude,
		SubmittedAt: timestamppb.New(dbSite.SubmittedAt.Time),
	}
}

// DBWithUserToProto maps sqlc DairySitesWithUser to proto GetDairySiteResponse
func DBWithUserToProto(dbSite db.DairySitesWithUser) *dairy_site.GetDairySiteResponse {
	return &dairy_site.GetDairySiteResponse{
		Id:          dbSite.ID,
		NameOfSite:  dbSite.NameOfSite,
		TodaysWork:  *dbSite.TodaysWork,
		Username:    *dbSite.Username,
		Fullname:    *dbSite.Fullname,
		Latitude:    *dbSite.Latitude,
		Longitude:   *dbSite.Longitude,
		SubmittedAt: timestamppb.New(dbSite.SubmittedAt.Time),
		CreatedAt:   timestamppb.New(dbSite.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(dbSite.UpdatedAt.Time),
	}
}
