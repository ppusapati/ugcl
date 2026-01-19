// =============================================================================
// services/csv_service.go - CSV processing and analysis service implementation
// =============================================================================
package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "p9e.in/ugcl/databridge/api/databridge"
	"p9e.in/ugcl/databridge/models"

	"github.com/google/uuid"
)

// CSVService implements ICSVService
type CSVService struct {
	fileService      IFileService
	validationService IValidationService
}

// NewCSVService creates a new CSV service instance
func NewCSVService(fileService IFileService, validationService IValidationService) ICSVService {
	return &CSVService{
		fileService:      fileService,
		validationService: validationService,
	}
}

// AnalyzeCSV performs comprehensive analysis of CSV data
func (s *CSVService) AnalyzeCSV(ctx context.Context, req *pb.AnalyzeCSVRequest) (*pb.CSVAnalysis, error) {
	startTime := time.Now()

	// Detect encoding if not UTF-8
	encoding, err := s.fileService.DetectCSVEncoding(req.CsvData)
	if err != nil {
		encoding = "UTF-8" // Default fallback
	}

	// Detect delimiter if not provided
	delimiter := req.Delimiter
	if delimiter == nil || *delimiter == "" {
		detectedDelimiter, err := s.fileService.DetectCSVDelimiter(req.CsvData)
		if err != nil {
			detectedDelimiter = "," // Default fallback
		}
		delimiter = &detectedDelimiter
	}

	// Parse CSV headers and sample data
	headers, err := s.ParseCSVHeaders(req.CsvData, *delimiter, req.HasHeader != nil && *req.HasHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV headers: %w", err)
	}

	// Parse CSV data for analysis
	parseResult, err := s.parseCSVData(req.CsvData, *delimiter, req.HasHeader != nil && *req.HasHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	// Infer column types
	columnAnalysis, err := s.InferColumnTypes(req.CsvData, headers, *delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to infer column types: %w", err)
	}

	// Build sample rows (first 5 data rows)
	sampleRows := make([]string, 0, 5)
	for i, row := range parseResult.Rows {
		if i >= 5 {
			break
		}
		if len(row) > 0 {
			rowData := make(map[string]string)
			for j, header := range headers {
				if j < len(row) {
					rowData[header] = row[j]
				}
			}
			jsonData, _ := json.Marshal(rowData)
			sampleRows = append(sampleRows, string(jsonData))
		}
	}

	// Convert column analysis to protobuf
	pbColumnAnalysis := make([]*pb.CSVColumnAnalysis, len(columnAnalysis))
	for i, col := range columnAnalysis {
		maxLength := col.MaxLength
		pbColumnAnalysis[i] = &pb.CSVColumnAnalysis{
			Header:        col.Header,
			NonEmptyCount: int32(col.UniqueCount),
			EmptyCount:    col.NullCount,
			InferredType:  col.InferredType,
			SampleValues:  col.SampleValues,
			MaxLength:     &maxLength,
			HasDuplicates: col.HasDuplicates,
		}
	}

	analysisTime := time.Since(startTime)

	return &pb.CSVAnalysis{
		Headers:         headers,
		TotalRows:       parseResult.TotalRows,
		DataRows:        parseResult.DataRows,
		ColumnAnalysis:  pbColumnAnalysis,
		SampleRows:      sampleRows,
		Encoding:        encoding,
		Delimiter:       *delimiter,
		HasHeader:       req.HasHeader != nil && *req.HasHeader,
		Errors:          parseResult.ParseErrors,
	}, nil
}

// ValidateCSVWithMapping validates CSV data against an import mapping
func (s *CSVService) ValidateCSVWithMapping(ctx context.Context, csvData []byte, mapping *pb.ImportMapping) (*models.ValidationResult, error) {
	startTime := time.Now()

	// Parse CSV headers
	var csvHeaders []string
	err := json.Unmarshal(mapping.CsvHeaders, &csvHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mapping CSV headers: %w", err)
	}

	// Parse field mappings
	var fieldMappings []*pb.FieldMapping
	err = json.Unmarshal(mapping.FieldMappings, &fieldMappings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse field mappings: %w", err)
	}

	// Parse CSV data
	parseResult, err := s.parseCSVData(csvData, ",", true) // Assuming comma delimiter and header
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	// Validate headers match
	headerErrors := s.validateHeaders(parseResult.Headers, csvHeaders)

	// Validate data rows
	var validationErrors []*models.ValidationError
	var warnings []*models.ValidationError
	validRows := int32(0)
	invalidRows := int32(0)

	for rowIndex, row := range parseResult.Rows {
		rowErrors := s.validateRow(row, parseResult.Headers, fieldMappings, int32(rowIndex+1))

		hasErrors := false
		for _, err := range rowErrors {
			if err.Severity == models.ValidationLevelError {
				validationErrors = append(validationErrors, err)
				hasErrors = true
			} else {
				warnings = append(warnings, err)
			}
		}

		if hasErrors {
			invalidRows++
		} else {
			validRows++
		}
	}

	// Add header errors
	for _, err := range headerErrors {
		validationErrors = append(validationErrors, err)
	}

	processingTime := time.Since(startTime).Milliseconds()

	return &models.ValidationResult{
		IsValid:          len(validationErrors) == 0,
		TotalRows:        parseResult.TotalRows,
		ValidRows:        validRows,
		InvalidRows:      invalidRows,
		Errors:           validationErrors,
		Warnings:         warnings,
		ProcessingTimeMs: processingTime,
	}, nil
}

// PreviewImport generates a preview of the import operation
func (s *CSVService) PreviewImport(ctx context.Context, req *pb.PreviewImportRequest) (*pb.PreviewImportResponse, error) {
	// Parse CSV data
	parseResult, err := s.parseCSVData(req.CsvData, ",", true) // Assuming comma delimiter and header
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	// Get the import mapping (this would typically be fetched from the mapping service)
	mapping := &pb.ImportMapping{
		Id:            req.MappingId,
		// This would be populated from the database
	}

	// Validate CSV with mapping
	validationResult, err := s.ValidateCSVWithMapping(ctx, req.CsvData, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to validate CSV: %w", err)
	}

	// Generate preview rows
	previewRowCount := int(req.PreviewRows)
	if previewRowCount <= 0 {
		previewRowCount = 10
	}
	if previewRowCount > len(parseResult.Rows) {
		previewRowCount = len(parseResult.Rows)
	}

	previewRows := make([]*pb.PreviewRow, previewRowCount)
	for i := 0; i < previewRowCount; i++ {
		row := parseResult.Rows[i]

		// Create original data map
		originalData := make(map[string]string)
		for j, header := range parseResult.Headers {
			if j < len(row) {
				originalData[header] = row[j]
			}
		}

		// Apply transformations (simplified version)
		transformedData := make(map[string]string)
		for key, value := range originalData {
			// Apply basic transformations
			transformedValue := s.applyBasicTransformation(value, "trim")
			transformedData[key] = transformedValue
		}

		// Find errors for this row
		var rowErrors []*pb.ImportError
		for _, err := range validationResult.Errors {
			if err.RowNumber == int32(i+1) {
				rowErrors = append(rowErrors, &pb.ImportError{
					RowNumber:    err.RowNumber,
					CsvField:     err.FieldName,
					DbColumn:     err.ColumnName,
					ErrorMessage: err.ErrorMessage,
					InvalidValue: err.Value,
					ErrorType:    err.ErrorType,
				})
			}
		}

		previewRows[i] = &pb.PreviewRow{
			RowNumber:       int32(i + 1),
			OriginalData:    originalData,
			TransformedData: transformedData,
			RowErrors:       rowErrors,
			IsValid:         len(rowErrors) == 0,
		}
	}

	// Convert validation errors to protobuf format
	pbErrors := make([]*pb.ImportError, len(validationResult.Errors))
	for i, err := range validationResult.Errors {
		pbErrors[i] = &pb.ImportError{
			RowNumber:    err.RowNumber,
			CsvField:     err.FieldName,
			DbColumn:     err.ColumnName,
			ErrorMessage: err.ErrorMessage,
			InvalidValue: err.Value,
			ErrorType:    err.ErrorType,
		}
	}

	warnings := make([]string, len(validationResult.Warnings))
	for i, warning := range validationResult.Warnings {
		warnings[i] = warning.ErrorMessage
	}

	return &pb.PreviewImportResponse{
		Headers:      parseResult.Headers,
		PreviewRows:  previewRows,
		Errors:       pbErrors,
		Warnings:     warnings,
		Mapping:      mapping,
	}, nil
}

// ParseCSVHeaders extracts headers from CSV data
func (s *CSVService) ParseCSVHeaders(csvData []byte, delimiter string, hasHeader bool) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	reader.Comma = rune(delimiter[0])

	if hasHeader {
		headers, err := reader.Read()
		if err != nil {
			return nil, err
		}
		return headers, nil
	}

	// If no header, read first row to determine column count
	firstRow, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Generate generic column names
	headers := make([]string, len(firstRow))
	for i := range headers {
		headers[i] = fmt.Sprintf("Column_%d", i+1)
	}

	return headers, nil
}

// InferColumnTypes analyzes CSV data to infer column data types
func (s *CSVService) InferColumnTypes(csvData []byte, headers []string, delimiter string) ([]*models.ColumnTypeInference, error) {
	parseResult, err := s.parseCSVData(csvData, delimiter, true)
	if err != nil {
		return nil, err
	}

	columnInferences := make([]*models.ColumnTypeInference, len(headers))

	for colIndex, header := range headers {
		inference := &models.ColumnTypeInference{
			Header:        header,
			SampleValues:  make([]string, 0, 5),
		}

		values := make([]string, 0)
		uniqueValues := make(map[string]bool)
		nullCount := int32(0)
		maxLength := int32(0)

		// Collect values for this column
		for _, row := range parseResult.Rows {
			if colIndex < len(row) {
				value := strings.TrimSpace(row[colIndex])
				if value == "" {
					nullCount++
				} else {
					values = append(values, value)
					uniqueValues[value] = true
					if len(value) > int(maxLength) {
						maxLength = int32(len(value))
					}
				}
			} else {
				nullCount++
			}
		}

		inference.NullCount = nullCount
		inference.UniqueCount = int32(len(uniqueValues))
		inference.MaxLength = maxLength
		inference.HasDuplicates = len(values) > len(uniqueValues)

		// Add sample values (first 5 unique)
		count := 0
		for value := range uniqueValues {
			if count >= 5 {
				break
			}
			inference.SampleValues = append(inference.SampleValues, value)
			count++
		}

		// Infer data type
		inference.InferredType, inference.Confidence = s.inferDataType(values)

		columnInferences[colIndex] = inference
	}

	return columnInferences, nil
}

// Helper methods

func (s *CSVService) parseCSVData(csvData []byte, delimiter string, hasHeader bool) (*models.CSVParseResult, error) {
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	reader.Comma = rune(delimiter[0])

	var headers []string
	var rows [][]string
	var parseErrors []string

	// Read headers if present
	if hasHeader {
		headerRow, err := reader.Read()
		if err != nil && err != io.EOF {
			return nil, err
		}
		headers = headerRow
	}

	// Read all data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			parseErrors = append(parseErrors, err.Error())
			continue
		}
		rows = append(rows, record)
	}

	// Generate headers if not present
	if !hasHeader && len(rows) > 0 {
		headers = make([]string, len(rows[0]))
		for i := range headers {
			headers[i] = fmt.Sprintf("Column_%d", i+1)
		}
	}

	return &models.CSVParseResult{
		Headers:     headers,
		Rows:        rows,
		TotalRows:   int32(len(rows)),
		DataRows:    int32(len(rows)),
		Delimiter:   delimiter,
		HasHeader:   hasHeader,
		ParseErrors: parseErrors,
	}, nil
}

func (s *CSVService) validateHeaders(csvHeaders, expectedHeaders []string) []*models.ValidationError {
	var errors []*models.ValidationError

	// Check if all expected headers are present
	csvHeaderMap := make(map[string]bool)
	for _, header := range csvHeaders {
		csvHeaderMap[header] = true
	}

	for _, expected := range expectedHeaders {
		if !csvHeaderMap[expected] {
			errors = append(errors, &models.ValidationError{
				RowNumber:    0,
				FieldName:    expected,
				ErrorType:    "missing_header",
				ErrorMessage: fmt.Sprintf("Required header '%s' not found in CSV", expected),
				Severity:     models.ValidationLevelError,
			})
		}
	}

	return errors
}

func (s *CSVService) validateRow(row, headers []string, mappings []*pb.FieldMapping, rowNumber int32) []*models.ValidationError {
	var errors []*models.ValidationError

	// Create row data map
	rowData := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			rowData[header] = row[i]
		}
	}

	// Validate each mapped field
	for _, mapping := range mappings {
		value, exists := rowData[mapping.CsvField]
		if !exists {
			errors = append(errors, &models.ValidationError{
				RowNumber:    rowNumber,
				FieldName:    mapping.CsvField,
				ColumnName:   mapping.DbColumn,
				ErrorType:    "missing_field",
				ErrorMessage: fmt.Sprintf("Field '%s' not found in row", mapping.CsvField),
				Severity:     models.ValidationLevelError,
			})
			continue
		}

		// Validate field value based on validations
		for _, validation := range mapping.Validations {
			if err := s.validateFieldByRule(value, validation, rowNumber, mapping.CsvField, mapping.DbColumn); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func (s *CSVService) validateFieldByRule(value string, rule *pb.ValidationRule, rowNumber int32, fieldName, columnName string) *models.ValidationError {
	switch rule.Type {
	case "required":
		if strings.TrimSpace(value) == "" {
			return &models.ValidationError{
				RowNumber:    rowNumber,
				FieldName:    fieldName,
				ColumnName:   columnName,
				Value:        value,
				ErrorType:    "required",
				ErrorMessage: fmt.Sprintf("Field '%s' is required but empty", fieldName),
				Severity:     models.ValidationLevelError,
			}
		}
	case "min_length":
		minLen, _ := strconv.Atoi(rule.Value)
		if len(value) < minLen {
			return &models.ValidationError{
				RowNumber:    rowNumber,
				FieldName:    fieldName,
				ColumnName:   columnName,
				Value:        value,
				ErrorType:    "min_length",
				ErrorMessage: fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, minLen),
				Severity:     models.ValidationLevelError,
			}
		}
	case "max_length":
		maxLen, _ := strconv.Atoi(rule.Value)
		if len(value) > maxLen {
			return &models.ValidationError{
				RowNumber:    rowNumber,
				FieldName:    fieldName,
				ColumnName:   columnName,
				Value:        value,
				ErrorType:    "max_length",
				ErrorMessage: fmt.Sprintf("Field '%s' must be at most %d characters", fieldName, maxLen),
				Severity:     models.ValidationLevelError,
			}
		}
	case "pattern":
		if matched, _ := regexp.MatchString(rule.Value, value); !matched {
			return &models.ValidationError{
				RowNumber:    rowNumber,
				FieldName:    fieldName,
				ColumnName:   columnName,
				Value:        value,
				ErrorType:    "pattern",
				ErrorMessage: fmt.Sprintf("Field '%s' does not match required pattern", fieldName),
				Severity:     models.ValidationLevelError,
			}
		}
	}

	return nil
}

func (s *CSVService) applyBasicTransformation(value, transformation string) string {
	switch transformation {
	case "trim":
		return strings.TrimSpace(value)
	case "uppercase":
		return strings.ToUpper(value)
	case "lowercase":
		return strings.ToLower(value)
	default:
		return value
	}
}

func (s *CSVService) inferDataType(values []string) (pb.DataType, float64) {
	if len(values) == 0 {
		return pb.DataType_DATA_TYPE_TEXT, 0.0
	}

	// Count different type patterns
	intCount := 0
	floatCount := 0
	boolCount := 0
	dateCount := 0
	uuidCount := 0

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		// Check for integer
		if _, err := strconv.Atoi(value); err == nil {
			intCount++
			continue
		}

		// Check for float
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			floatCount++
			continue
		}

		// Check for boolean
		if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" ||
		   value == "1" || value == "0" {
			boolCount++
			continue
		}

		// Check for UUID
		if _, err := uuid.Parse(value); err == nil {
			uuidCount++
			continue
		}

		// Check for date patterns
		datePatterns := []string{
			"2006-01-02",
			"01/02/2006",
			"2006-01-02 15:04:05",
			"01/02/2006 15:04:05",
		}
		for _, pattern := range datePatterns {
			if _, err := time.Parse(pattern, value); err == nil {
				dateCount++
				break
			}
		}
	}

	total := len(values)

	// Determine type based on highest match percentage
	if uuidCount > 0 && float64(uuidCount)/float64(total) > 0.8 {
		return pb.DataType_DATA_TYPE_UUID, float64(uuidCount) / float64(total)
	}
	if intCount > 0 && float64(intCount)/float64(total) > 0.8 {
		return pb.DataType_DATA_TYPE_INTEGER, float64(intCount) / float64(total)
	}
	if floatCount > 0 && float64(floatCount+intCount)/float64(total) > 0.8 {
		return pb.DataType_DATA_TYPE_DECIMAL, float64(floatCount+intCount) / float64(total)
	}
	if boolCount > 0 && float64(boolCount)/float64(total) > 0.8 {
		return pb.DataType_DATA_TYPE_BOOLEAN, float64(boolCount) / float64(total)
	}
	if dateCount > 0 && float64(dateCount)/float64(total) > 0.8 {
		return pb.DataType_DATA_TYPE_TIMESTAMP, float64(dateCount) / float64(total)
	}

	// Default to text
	return pb.DataType_DATA_TYPE_TEXT, 1.0
}