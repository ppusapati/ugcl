package mappers

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "p9e.in/ugcl/vendors/api/v2/vendor"
	db "p9e.in/ugcl/vendors/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

func ProtoToDBVendor(p *pb.Vendor) (*db.Vendor, error) {
	var metadataBytes []byte
	if p.Metadata != nil {
		b, err := json.Marshal(p.Metadata.AsMap())
		if err != nil {
			return nil, err
		}
		metadataBytes = b
	}

	return &db.Vendor{
		ID:               p.Id,
		CompanyName:      p.CompanyName,
		CompanyType:      strPtrOrNil(p.CompanyType),
		Gst:              strPtrOrNil(p.Gst),
		Pan:              strPtrOrNil(p.Pan),
		VendorCategory:   strPtrOrNil(p.VendorCategory),
		PersonID:         p.PersonId,
		PaymentTerms:     strPtrOrNil(p.PaymentTerms),
		CreditPeriodDays: int32PtrOrNil(p.CreditPeriodDays),
		BankName:         strPtrOrNil(p.BankName),
		AccountNumber:    strPtrOrNil(p.AccountNumber),
		Ifsc:             strPtrOrNil(p.Ifsc),
		Rating:           int32PtrOrNil(p.Rating),
		IsBlacklisted:    boolPtrOrNil(p.IsBlacklisted),
		Contracts:        p.Contracts,
		PurchaseOrders:   p.PurchaseOrders,
		Status:           strPtrOrNil(p.Status),
		Metadata:         metadataBytes,
		CreatedAt:        timestampProtoToPg(p.CreatedAt),
		UpdatedAt:        timestampProtoToPg(p.UpdatedAt),
	}, nil
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32PtrOrNil(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

func boolPtrOrNil(b bool) *bool {
	return &b
}

func timestampProtoToPg(t *timestamppb.Timestamp) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t.AsTime(), Valid: true}
}

func DBToProtoVendor(db *db.Vendor) (*pb.Vendor, error) {
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

	return &pb.Vendor{
		Id:               db.ID,
		CompanyName:      db.CompanyName,
		CompanyType:      strValOrEmpty(db.CompanyType),
		Gst:              strValOrEmpty(db.Gst),
		Pan:              strValOrEmpty(db.Pan),
		VendorCategory:   strValOrEmpty(db.VendorCategory),
		PersonId:         db.PersonID,
		PaymentTerms:     strValOrEmpty(db.PaymentTerms),
		CreditPeriodDays: int32ValOrZero(db.CreditPeriodDays),
		BankName:         strValOrEmpty(db.BankName),
		AccountNumber:    strValOrEmpty(db.AccountNumber),
		Ifsc:             strValOrEmpty(db.Ifsc),
		Rating:           int32ValOrZero(db.Rating),
		IsBlacklisted:    boolValOrFalse(db.IsBlacklisted),
		Contracts:        db.Contracts,
		PurchaseOrders:   db.PurchaseOrders,
		Status:           strValOrEmpty(db.Status),
		Metadata:         metaStruct,
		CreatedAt:        pgToTimestampProto(db.CreatedAt),
		UpdatedAt:        pgToTimestampProto(db.UpdatedAt),
	}, nil
}

func strValOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int32ValOrZero(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func boolValOrFalse(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func pgToTimestampProto(ts pgtype.Timestamp) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}

func ApplyFieldMask(existing *db.Vendor, updated *pb.Vendor, mask *fieldmaskpb.FieldMask) *db.Vendor {
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
		case "vendor_category":
			existing.VendorCategory = strPtrOrNil(updated.VendorCategory)
		case "payment_terms":
			existing.PaymentTerms = strPtrOrNil(updated.PaymentTerms)
		case "credit_period_days":
			existing.CreditPeriodDays = int32PtrOrNil(updated.CreditPeriodDays)
		case "bank_name":
			existing.BankName = strPtrOrNil(updated.BankName)
		case "account_number":
			existing.AccountNumber = strPtrOrNil(updated.AccountNumber)
		case "ifsc":
			existing.Ifsc = strPtrOrNil(updated.Ifsc)
		case "rating":
			existing.Rating = int32PtrOrNil(updated.Rating)
		case "is_blacklisted":
			existing.IsBlacklisted = boolPtrOrNil(updated.IsBlacklisted)
		case "contracts":
			existing.Contracts = updated.Contracts
		case "purchase_orders":
			existing.PurchaseOrders = updated.PurchaseOrders
		case "status":
			existing.Status = strPtrOrNil(updated.Status)
		case "metadata":
			b, _ := updated.Metadata.MarshalJSON()
			existing.Metadata = b
		}
	}
	return existing
}
