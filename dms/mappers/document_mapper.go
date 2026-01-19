package mappers

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	db "p9e.in/ugcl/dms/db/generated"
	pb "p9e.in/ugcl/dms/api/v1"
)

// DocumentDBToProto converts database Document to proto Document
func DocumentDBToProto(dbDoc *db.Document) (*pb.Document, error) {
	if dbDoc == nil {
		return nil, nil
	}

	metadata, err := JSONRawMessageToMap(dbDoc.Metadata)
	if err != nil {
		return nil, err
	}

	doc := &pb.Document{
		Id:               dbDoc.ID.String(),
		TenantId:         dbDoc.TenantID.String(),
		DocumentType:     DocumentTypeDBToProto(dbDoc.DocumentType),
		DocumentCategory: DocumentCategoryDBToProto(dbDoc.DocumentCategory),
		FileName:         dbDoc.FileName,
		OriginalName:     dbDoc.OriginalName,
		MimeType:         dbDoc.MimeType,
		SizeBytes:        dbDoc.SizeBytes,
		CompressedSize:   dbDoc.CompressedSize,
		Checksum:         dbDoc.Checksum,
		StoragePath:      dbDoc.StoragePath,
		PageCount:        dbDoc.PageCount,
		WordCount:        dbDoc.WordCount,
		ProcessingStatus: ProcessingStatusDBToProto(dbDoc.ProcessingStatus),
		CompressionRatio: dbDoc.CompressionRatio.Float64,
		OcrStatus:        OCRStatusDBToProto(dbDoc.OcrStatus),
		OcrConfidence:    dbDoc.OcrConfidence.Float64,
		VirusScanStatus:  VirusScanStatusDBToProto(dbDoc.VirusScanStatus),
		HasWatermark:     dbDoc.HasWatermark,
		UploadedBy:       dbDoc.UploadedBy.String(),
		Tags:             PostgresArrayToStringArray(dbDoc.Tags),
		Metadata:         metadata,
		IsExpired:        dbDoc.IsExpired,
		CreatedAt:        TimeToTimestamp(dbDoc.CreatedAt),
		UpdatedAt:        TimeToTimestamp(dbDoc.UpdatedAt),
		CreatedBy:        dbDoc.CreatedBy.String(),
		UpdatedBy:        dbDoc.UpdatedBy.String(),
		IsDeleted:        dbDoc.IsDeleted,
	}

	// Optional fields
	if dbDoc.OwnerEntityType != nil {
		doc.OwnerEntityType = wrapperspb.String(*dbDoc.OwnerEntityType)
	}
	if dbDoc.OwnerEntityID != nil {
		doc.OwnerEntityId = wrapperspb.String(dbDoc.OwnerEntityID.String())
	}
	if dbDoc.DivisionID != nil {
		doc.DivisionId = wrapperspb.String(dbDoc.DivisionID.String())
	}
	if dbDoc.BranchID != nil {
		doc.BranchId = wrapperspb.String(dbDoc.BranchID.String())
	}
	if dbDoc.DepartmentID != nil {
		doc.DepartmentId = wrapperspb.String(dbDoc.DepartmentID.String())
	}
	if dbDoc.Title != nil {
		doc.Title = wrapperspb.String(*dbDoc.Title)
	}
	if dbDoc.Description != nil {
		doc.Description = wrapperspb.String(*dbDoc.Description)
	}
	if dbDoc.FileExtension != nil {
		doc.FileExtension = wrapperspb.String(*dbDoc.FileExtension)
	}
	if dbDoc.CompressedPath != nil {
		doc.CompressedPath = wrapperspb.String(*dbDoc.CompressedPath)
	}
	if dbDoc.ThumbnailPath != nil {
		doc.ThumbnailPath = wrapperspb.String(*dbDoc.ThumbnailPath)
	}
	if dbDoc.PreviewPath != nil {
		doc.PreviewPath = wrapperspb.String(*dbDoc.PreviewPath)
	}
	if dbDoc.ExtractedText != nil {
		doc.ExtractedText = wrapperspb.String(*dbDoc.ExtractedText)
	}
	if dbDoc.Language != nil {
		doc.Language = wrapperspb.String(*dbDoc.Language)
	}
	if dbDoc.Author != nil {
		doc.Author = wrapperspb.String(*dbDoc.Author)
	}
	if dbDoc.Subject != nil {
		doc.Subject = wrapperspb.String(*dbDoc.Subject)
	}
	if dbDoc.Keywords != nil {
		doc.Keywords = wrapperspb.String(*dbDoc.Keywords)
	}
	if dbDoc.ProcessingError != nil {
		doc.ProcessingError = wrapperspb.String(*dbDoc.ProcessingError)
	}
	if dbDoc.ProcessedAt != nil {
		doc.ProcessedAt = TimeToTimestamp(*dbDoc.ProcessedAt)
	}
	if dbDoc.CompressionType != nil {
		doc.CompressionType = wrapperspb.String(*dbDoc.CompressionType)
	}
	if dbDoc.OcrProcessedAt != nil {
		doc.OcrProcessedAt = TimeToTimestamp(*dbDoc.OcrProcessedAt)
	}
	if dbDoc.VirusScanResult != nil {
		doc.VirusScanResult = wrapperspb.String(*dbDoc.VirusScanResult)
	}
	if dbDoc.VirusScannedAt != nil {
		doc.VirusScanedAt = TimeToTimestamp(*dbDoc.VirusScannedAt)
	}
	if dbDoc.ExpiresAt != nil {
		doc.ExpiresAt = TimeToTimestamp(*dbDoc.ExpiresAt)
	}
	if dbDoc.DeletedAt != nil {
		doc.DeletedAt = TimeToTimestamp(*dbDoc.DeletedAt)
	}
	if dbDoc.DeletedBy != nil {
		doc.DeletedBy = wrapperspb.String(dbDoc.DeletedBy.String())
	}

	return doc, nil
}

// Enum conversion helpers

func DocumentTypeDBToProto(dt db.DocumentType) pb.DocumentType {
	switch dt {
	case "PDF":
		return pb.DocumentType_DOCUMENT_TYPE_PDF
	case "IMAGE":
		return pb.DocumentType_DOCUMENT_TYPE_IMAGE
	case "VIDEO":
		return pb.DocumentType_DOCUMENT_TYPE_VIDEO
	case "AUDIO":
		return pb.DocumentType_DOCUMENT_TYPE_AUDIO
	case "SPREADSHEET":
		return pb.DocumentType_DOCUMENT_TYPE_SPREADSHEET
	case "PRESENTATION":
		return pb.DocumentType_DOCUMENT_TYPE_PRESENTATION
	case "TEXT":
		return pb.DocumentType_DOCUMENT_TYPE_TEXT
	case "ARCHIVE":
		return pb.DocumentType_DOCUMENT_TYPE_ARCHIVE
	case "OTHER":
		return pb.DocumentType_DOCUMENT_TYPE_OTHER
	default:
		return pb.DocumentType_DOCUMENT_TYPE_UNSPECIFIED
	}
}

func DocumentTypeProtoToDB(dt pb.DocumentType) db.DocumentType {
	switch dt {
	case pb.DocumentType_DOCUMENT_TYPE_PDF:
		return "PDF"
	case pb.DocumentType_DOCUMENT_TYPE_IMAGE:
		return "IMAGE"
	case pb.DocumentType_DOCUMENT_TYPE_VIDEO:
		return "VIDEO"
	case pb.DocumentType_DOCUMENT_TYPE_AUDIO:
		return "AUDIO"
	case pb.DocumentType_DOCUMENT_TYPE_SPREADSHEET:
		return "SPREADSHEET"
	case pb.DocumentType_DOCUMENT_TYPE_PRESENTATION:
		return "PRESENTATION"
	case pb.DocumentType_DOCUMENT_TYPE_TEXT:
		return "TEXT"
	case pb.DocumentType_DOCUMENT_TYPE_ARCHIVE:
		return "ARCHIVE"
	case pb.DocumentType_DOCUMENT_TYPE_OTHER:
		return "OTHER"
	default:
		return "OTHER"
	}
}

func DocumentCategoryDBToProto(dc db.DocumentCategory) pb.DocumentCategory {
	switch dc {
	case "CONTRACT":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_CONTRACT
	case "INVOICE":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_INVOICE
	case "REPORT":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_REPORT
	case "MEMO":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_MEMO
	case "LETTER":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_LETTER
	case "FORM":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_FORM
	case "POLICY":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_POLICY
	case "PROCEDURE":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_PROCEDURE
	case "MANUAL":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_MANUAL
	case "PRESENTATION":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_PRESENTATION
	case "DRAWING":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_DRAWING
	case "PHOTO":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_PHOTO
	case "VIDEO":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_VIDEO
	case "AUDIO":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_AUDIO
	case "OTHER":
		return pb.DocumentCategory_DOCUMENT_CATEGORY_OTHER
	default:
		return pb.DocumentCategory_DOCUMENT_CATEGORY_UNSPECIFIED
	}
}

func DocumentCategoryProtoToDB(dc pb.DocumentCategory) db.DocumentCategory {
	switch dc {
	case pb.DocumentCategory_DOCUMENT_CATEGORY_CONTRACT:
		return "CONTRACT"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_INVOICE:
		return "INVOICE"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_REPORT:
		return "REPORT"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_MEMO:
		return "MEMO"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_LETTER:
		return "LETTER"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_FORM:
		return "FORM"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_POLICY:
		return "POLICY"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_PROCEDURE:
		return "PROCEDURE"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_MANUAL:
		return "MANUAL"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_PRESENTATION:
		return "PRESENTATION"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_DRAWING:
		return "DRAWING"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_PHOTO:
		return "PHOTO"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_VIDEO:
		return "VIDEO"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_AUDIO:
		return "AUDIO"
	case pb.DocumentCategory_DOCUMENT_CATEGORY_OTHER:
		return "OTHER"
	default:
		return "OTHER"
	}
}

func ProcessingStatusDBToProto(ps db.ProcessingStatus) pb.ProcessingStatus {
	switch ps {
	case "PENDING":
		return pb.ProcessingStatus_PROCESSING_STATUS_PENDING
	case "PROCESSING":
		return pb.ProcessingStatus_PROCESSING_STATUS_PROCESSING
	case "COMPLETED":
		return pb.ProcessingStatus_PROCESSING_STATUS_COMPLETED
	case "FAILED":
		return pb.ProcessingStatus_PROCESSING_STATUS_FAILED
	default:
		return pb.ProcessingStatus_PROCESSING_STATUS_UNSPECIFIED
	}
}

func OCRStatusDBToProto(os db.OcrStatus) pb.OCRStatus {
	switch os {
	case "PENDING":
		return pb.OCRStatus_OCR_STATUS_PENDING
	case "PROCESSING":
		return pb.OCRStatus_OCR_STATUS_PROCESSING
	case "COMPLETED":
		return pb.OCRStatus_OCR_STATUS_COMPLETED
	case "FAILED":
		return pb.OCRStatus_OCR_STATUS_FAILED
	case "SKIPPED":
		return pb.OCRStatus_OCR_STATUS_SKIPPED
	default:
		return pb.OCRStatus_OCR_STATUS_UNSPECIFIED
	}
}

func VirusScanStatusDBToProto(vs db.VirusScanStatus) pb.VirusScanStatus {
	switch vs {
	case "PENDING":
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_PENDING
	case "SCANNING":
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_SCANNING
	case "CLEAN":
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_CLEAN
	case "INFECTED":
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_INFECTED
	case "FAILED":
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_FAILED
	default:
		return pb.VirusScanStatus_VIRUS_SCAN_STATUS_UNSPECIFIED
	}
}
