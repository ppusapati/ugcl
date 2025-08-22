package mappers

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/vendors/api/v2/contractor"
	db "p9e.in/ugcl/vendors/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

func ProtoToDBContractor(p *pb.Contractor) (*db.Contractor, error) {
	var metadataBytes []byte
	if p.Metadata != nil {
		b, err := json.Marshal(p.Metadata.AsMap())
		if err != nil {
			return nil, err
		}
		metadataBytes = b
	}

	return &db.Contractor{
		ID:                p.Id,
		CompanyName:       p.CompanyName,
		CompanyType:       strPtrOrNil(p.CompanyType),
		Gst:               strPtrOrNil(p.Gst),
		Pan:               strPtrOrNil(p.Pan),
		Category:          strPtrOrNil(p.Category),
		PersonID:          p.PersonId,
		AssociatedProject: p.AssociatedProject,
		WorkingSite:       p.WorkingSite,
		ContractStartDate: timestampProtoToPg(p.ContractStartDate),
		ContractEndDate:   timestampProtoToPg(p.ContractEndDate),
		Status:            strPtrOrNil(p.Status),
		Metadata:          metadataBytes,
		CreatedAt:         timestampProtoToPg(p.CreatedAt),
		UpdatedAt:         timestampProtoToPg(p.UpdatedAt),
	}, nil
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timestampProtoToPg(t *timestamppb.Timestamp) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t.AsTime(), Valid: true}
}

func DBToProtoContractor(db *db.Contractor) (*pb.Contractor, error) {
	var metaStruct *structpb.Struct
	if len(db.Metadata) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(db.Metadata, &m); err != nil {
			return nil, err
		}
		s, err := structpb.NewStruct(m)
		if err != nil {
			return nil, err
		}
		metaStruct = s
	}

	return &pb.Contractor{
		Id:                db.ID,
		CompanyName:       db.CompanyName,
		CompanyType:       strValOrEmpty(db.CompanyType),
		Gst:               strValOrEmpty(db.Gst),
		Pan:               strValOrEmpty(db.Pan),
		Category:          strValOrEmpty(db.Category),
		PersonId:          db.PersonID,
		AssociatedProject: db.AssociatedProject,
		WorkingSite:       db.WorkingSite,
		ContractStartDate: pgToTimestampProto(db.ContractStartDate),
		ContractEndDate:   pgToTimestampProto(db.ContractEndDate),
		Status:            strValOrEmpty(db.Status),
		Metadata:          metaStruct,
		CreatedAt:         pgToTimestampProto(db.CreatedAt),
		UpdatedAt:         pgToTimestampProto(db.UpdatedAt),
	}, nil
}

func strValOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func pgToTimestampProto(ts pgtype.Timestamp) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}

func ApplyFieldMask(existing *db.Contractor, updated *pb.Contractor, mask *fieldmaskpb.FieldMask) *db.Contractor {
	for _, path := range mask.Paths {
		switch path {
		case "company_name":
			existing.CompanyName = updated.CompanyName
		case "company_type":
			existing.CompanyType = strPtrOrNil(updated.CompanyType)
		case "gst":
			existing.Gst = strPtrOrNil(updated.Gst)
		case "pan":
			existing.Pan = strPtrOrNil(updated.Pan)
		case "category":
			existing.Category = strPtrOrNil(updated.Category)
		case "associated_project":
			existing.AssociatedProject = updated.AssociatedProject
		case "working_site":
			existing.WorkingSite = updated.WorkingSite
		case "contract_start_date":
			existing.ContractStartDate = timestampProtoToPg(updated.ContractStartDate)
		case "contract_end_date":
			existing.ContractEndDate = timestampProtoToPg(updated.ContractEndDate)
		case "status":
			existing.Status = strPtrOrNil(updated.Status)
		case "metadata":
			b, _ := updated.Metadata.MarshalJSON()
			existing.Metadata = b
		}
	}
	return existing
}
