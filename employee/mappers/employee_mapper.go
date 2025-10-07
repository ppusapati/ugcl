package mappers

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "p9e.in/ugcl/employee/api/v2/employee"
	db "p9e.in/ugcl/employee/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

// ProtoToDBEmployee converts proto Employee to DB Employee model
func ProtoToDBEmployee(p *pb.Employee) (*db.Employee, error) {
	var metadataBytes []byte
	if p.Metadata != nil {
		b, err := json.Marshal(p.Metadata.AsMap())
		if err != nil {
			return nil, err
		}
		metadataBytes = b
	}

	return &db.Employee{
		Uuid:                     uuidStrToPg(p.Id),
		EmployeeCode:             p.EmployeeCode,
		UserID:                   uuidStrToPg(p.UserId),
		DivisionID:               uuidWrapperToPg(p.DivisionId),
		BranchID:                 uuidWrapperToPg(p.BranchId),
		DepartmentID:             uuidWrapperToPg(p.DepartmentId),
		Designation:              p.Designation,
		JobTitle:                 strWrapperToPgText(p.JobTitle),
		JobGrade:                 strWrapperToPgText(p.JobGrade),
		EmployeeType:             strWrapperToPgText(p.EmployeeType),
		ManagerID:                uuidWrapperToPg(p.ManagerId),
		DepartmentHeadID:         uuidWrapperToPg(p.DepartmentHeadId),
		DateOfJoining:            timestampProtoToPg(p.DateOfJoining),
		DateOfConfirmation:       timestampProtoToPg(p.DateOfConfirmation),
		DateOfLeaving:            timestampProtoToPg(p.DateOfLeaving),
		Pan:                      strWrapperToPgText(p.Pan),
		Aadhaar:                  strWrapperToPgText(p.Aadhaar),
		Uan:                      strWrapperToPgText(p.Uan),
		EsicNumber:               strWrapperToPgText(p.EsicNumber),
		BankName:                 strWrapperToPgText(p.BankName),
		BankAccountNumber:        strWrapperToPgText(p.BankAccountNumber),
		BankIfsc:                 strWrapperToPgText(p.BankIfsc),
		CurrentAddress:           strWrapperToPgText(p.CurrentAddress),
		PermanentAddress:         strWrapperToPgText(p.PermanentAddress),
		EmergencyContactName:     strWrapperToPgText(p.EmergencyContactName),
		EmergencyContactPhone:    strWrapperToPgText(p.EmergencyContactPhone),
		EmergencyContactRelation: strWrapperToPgText(p.EmergencyContactRelation),
		WorkLocation:             strWrapperToPgText(p.WorkLocation),
		OfficePhone:              strWrapperToPgText(p.OfficePhone),
		Extension:                strWrapperToPgText(p.Extension),
		Status:                   strToPgText(p.Status),
		TerminationReason:        strWrapperToPgText(p.TerminationReason),
		Skills:                   p.Skills,
		Certifications:           p.Certifications,
		HighestQualification:     strWrapperToPgText(p.HighestQualification),
		Metadata:                 metadataBytes,
		CreatedAt:                timestampProtoToPg(p.CreatedAt),
		UpdatedAt:                timestampProtoToPg(p.UpdatedAt),
		CreatedBy:                uuidWrapperToPg(p.CreatedBy),
		UpdatedBy:                uuidWrapperToPg(p.UpdatedBy),
	}, nil
}

// DBToProtoEmployee converts DB Employee to proto Employee model
func DBToProtoEmployee(d *db.Employee) (*pb.Employee, error) {
	var metaStruct *structpb.Struct
	if len(d.Metadata) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(d.Metadata, &m); err != nil {
			return nil, err
		}
		s, err := structpb.NewStruct(m)
		if err != nil {
			return nil, err
		}
		metaStruct = s
	}

	return &pb.Employee{
		Id:                       pgUuidToStr(d.Uuid),
		EmployeeCode:             d.EmployeeCode,
		UserId:                   pgUuidToStr(d.UserID),
		DivisionId:               pgUuidToStrWrapper(d.DivisionID),
		BranchId:                 pgUuidToStrWrapper(d.BranchID),
		DepartmentId:             pgUuidToStrWrapper(d.DepartmentID),
		Designation:              d.Designation,
		JobTitle:                 pgTextToStrWrapper(d.JobTitle),
		JobGrade:                 pgTextToStrWrapper(d.JobGrade),
		EmployeeType:             pgTextToStrWrapper(d.EmployeeType),
		ManagerId:                pgUuidToStrWrapper(d.ManagerID),
		DepartmentHeadId:         pgUuidToStrWrapper(d.DepartmentHeadID),
		DateOfJoining:            pgToTimestampProto(d.DateOfJoining),
		DateOfConfirmation:       pgToTimestampProto(d.DateOfConfirmation),
		DateOfLeaving:            pgToTimestampProto(d.DateOfLeaving),
		Pan:                      pgTextToStrWrapper(d.Pan),
		Aadhaar:                  pgTextToStrWrapper(d.Aadhaar),
		Uan:                      pgTextToStrWrapper(d.Uan),
		EsicNumber:               pgTextToStrWrapper(d.EsicNumber),
		BankName:                 pgTextToStrWrapper(d.BankName),
		BankAccountNumber:        pgTextToStrWrapper(d.BankAccountNumber),
		BankIfsc:                 pgTextToStrWrapper(d.BankIfsc),
		CurrentAddress:           pgTextToStrWrapper(d.CurrentAddress),
		PermanentAddress:         pgTextToStrWrapper(d.PermanentAddress),
		EmergencyContactName:     pgTextToStrWrapper(d.EmergencyContactName),
		EmergencyContactPhone:    pgTextToStrWrapper(d.EmergencyContactPhone),
		EmergencyContactRelation: pgTextToStrWrapper(d.EmergencyContactRelation),
		WorkLocation:             pgTextToStrWrapper(d.WorkLocation),
		OfficePhone:              pgTextToStrWrapper(d.OfficePhone),
		Extension:                pgTextToStrWrapper(d.Extension),
		Status:                   pgTextToStr(d.Status),
		TerminationReason:        pgTextToStrWrapper(d.TerminationReason),
		Skills:                   d.Skills,
		Certifications:           d.Certifications,
		HighestQualification:     pgTextToStrWrapper(d.HighestQualification),
		Metadata:                 metaStruct,
		CreatedAt:                pgToTimestampProto(d.CreatedAt),
		UpdatedAt:                pgToTimestampProto(d.UpdatedAt),
		CreatedBy:                pgUuidToStrWrapper(d.CreatedBy),
		UpdatedBy:                pgUuidToStrWrapper(d.UpdatedBy),
	}, nil
}

// ApplyFieldMask applies field mask updates to existing employee
func ApplyFieldMask(existing *db.Employee, updated *pb.Employee, mask *fieldmaskpb.FieldMask) *db.Employee {
	for _, path := range mask.Paths {
		switch path {
		case "employee_code":
			existing.EmployeeCode = updated.EmployeeCode
		case "division_id":
			existing.DivisionID = uuidWrapperToPg(updated.DivisionId)
		case "branch_id":
			existing.BranchID = uuidWrapperToPg(updated.BranchId)
		case "department_id":
			existing.DepartmentID = uuidWrapperToPg(updated.DepartmentId)
		case "designation":
			existing.Designation = updated.Designation
		case "job_title":
			existing.JobTitle = strWrapperToPgText(updated.JobTitle)
		case "job_grade":
			existing.JobGrade = strWrapperToPgText(updated.JobGrade)
		case "employee_type":
			existing.EmployeeType = strWrapperToPgText(updated.EmployeeType)
		case "manager_id":
			existing.ManagerID = uuidWrapperToPg(updated.ManagerId)
		case "department_head_id":
			existing.DepartmentHeadID = uuidWrapperToPg(updated.DepartmentHeadId)
		case "date_of_joining":
			existing.DateOfJoining = timestampProtoToPg(updated.DateOfJoining)
		case "date_of_confirmation":
			existing.DateOfConfirmation = timestampProtoToPg(updated.DateOfConfirmation)
		case "date_of_leaving":
			existing.DateOfLeaving = timestampProtoToPg(updated.DateOfLeaving)
		case "pan":
			existing.Pan = strWrapperToPgText(updated.Pan)
		case "aadhaar":
			existing.Aadhaar = strWrapperToPgText(updated.Aadhaar)
		case "uan":
			existing.Uan = strWrapperToPgText(updated.Uan)
		case "esic_number":
			existing.EsicNumber = strWrapperToPgText(updated.EsicNumber)
		case "bank_name":
			existing.BankName = strWrapperToPgText(updated.BankName)
		case "bank_account_number":
			existing.BankAccountNumber = strWrapperToPgText(updated.BankAccountNumber)
		case "bank_ifsc":
			existing.BankIfsc = strWrapperToPgText(updated.BankIfsc)
		case "current_address":
			existing.CurrentAddress = strWrapperToPgText(updated.CurrentAddress)
		case "permanent_address":
			existing.PermanentAddress = strWrapperToPgText(updated.PermanentAddress)
		case "emergency_contact_name":
			existing.EmergencyContactName = strWrapperToPgText(updated.EmergencyContactName)
		case "emergency_contact_phone":
			existing.EmergencyContactPhone = strWrapperToPgText(updated.EmergencyContactPhone)
		case "emergency_contact_relation":
			existing.EmergencyContactRelation = strWrapperToPgText(updated.EmergencyContactRelation)
		case "work_location":
			existing.WorkLocation = strWrapperToPgText(updated.WorkLocation)
		case "office_phone":
			existing.OfficePhone = strWrapperToPgText(updated.OfficePhone)
		case "extension":
			existing.Extension = strWrapperToPgText(updated.Extension)
		case "status":
			existing.Status = strToPgText(updated.Status)
		case "termination_reason":
			existing.TerminationReason = strWrapperToPgText(updated.TerminationReason)
		case "skills":
			existing.Skills = updated.Skills
		case "certifications":
			existing.Certifications = updated.Certifications
		case "highest_qualification":
			existing.HighestQualification = strWrapperToPgText(updated.HighestQualification)
		case "metadata":
			if updated.Metadata != nil {
				b, _ := updated.Metadata.MarshalJSON()
				existing.Metadata = b
			}
		}
	}
	return existing
}

// Helper functions for UUID conversions
func uuidStrToPg(s string) pgtype.UUID {
	var uuid pgtype.UUID
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	_ = uuid.Scan(s)
	return uuid
}

func uuidWrapperToPg(w *wrapperspb.StringValue) pgtype.UUID {
	if w == nil || w.Value == "" {
		return pgtype.UUID{Valid: false}
	}
	var uuid pgtype.UUID
	_ = uuid.Scan(w.Value)
	return uuid
}

func pgUuidToStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	var s string
	u.AssignTo(&s)
	return s
}

func pgUuidToStrWrapper(u pgtype.UUID) *wrapperspb.StringValue {
	if !u.Valid {
		return nil
	}
	var s string
	u.AssignTo(&s)
	return wrapperspb.String(s)
}

// Helper functions for timestamp conversions
func timestampProtoToPg(t *timestamppb.Timestamp) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t.AsTime(), Valid: true}
}

func pgToTimestampProto(ts pgtype.Timestamp) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}

// Helper functions for string conversions
func strWrapperToPgText(w *wrapperspb.StringValue) pgtype.Text {
	if w == nil || w.Value == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: w.Value, Valid: true}
}

func strToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTextToStrWrapper(t pgtype.Text) *wrapperspb.StringValue {
	if !t.Valid {
		return nil
	}
	return wrapperspb.String(t.String)
}

func pgTextToStr(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// PgUuidToStr converts pgtype.UUID to string (exported for use in other packages)
func PgUuidToStr(u pgtype.UUID) string {
	return pgUuidToStr(u)
}
