// =============================================================================
// services/file_service.go - File handling service implementation
// =============================================================================
package services

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// FileService implements IFileService
type FileService struct {
	maxFileSize   int64
	allowedTypes  []string
}

// NewFileService creates a new file service instance
func NewFileService() IFileService {
	return &FileService{
		maxFileSize: 100 * 1024 * 1024, // 100MB
		allowedTypes: []string{".csv", ".txt"},
	}
}

// ValidateFileSize checks if file size is within allowed limits
func (s *FileService) ValidateFileSize(size int64) error {
	if size > s.maxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes", size, s.maxFileSize)
	}
	if size == 0 {
		return fmt.Errorf("file is empty")
	}
	return nil
}

// ValidateFileType checks if file type is allowed
func (s *FileService) ValidateFileType(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedType := range s.allowedTypes {
		if ext == allowedType {
			return nil
		}
	}

	return fmt.Errorf("file type '%s' is not allowed. Allowed types: %v", ext, s.allowedTypes)
}

// CalculateFileHash calculates SHA-256 hash of file data
func (s *FileService) CalculateFileHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// DetectCSVEncoding detects the encoding of CSV data
func (s *FileService) DetectCSVEncoding(data []byte) (string, error) {
	// Check if the data is valid UTF-8
	if utf8.Valid(data) {
		return "UTF-8", nil
	}

	// Try common encodings
	encodings := []struct {
		name     string
		encoding *charmap.Charmap
	}{
		{"ISO-8859-1", charmap.ISO8859_1},
		{"Windows-1252", charmap.Windows1252},
		{"UTF-16", nil}, // Special handling needed for UTF-16
	}

	for _, enc := range encodings {
		if enc.encoding == nil {
			continue
		}

		// Try to decode with this encoding
		decoder := enc.encoding.NewDecoder()
		decoded, _, err := transform.Bytes(decoder, data[:min(1000, len(data))]) // Test first 1000 bytes
		if err == nil && utf8.Valid(decoded) {
			return enc.name, nil
		}
	}

	// Default to UTF-8 if detection fails
	return "UTF-8", fmt.Errorf("could not detect encoding, defaulting to UTF-8")
}

// DetectCSVDelimiter detects the most likely delimiter used in CSV data
func (s *FileService) DetectCSVDelimiter(data []byte) (string, error) {
	// Common delimiters to test
	delimiters := []string{",", ";", "\t", "|"}

	// Take first few lines for analysis
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return ",", fmt.Errorf("insufficient data to detect delimiter, defaulting to comma")
	}

	// Test first 5 lines or all lines if fewer
	testLines := lines[:min(5, len(lines))]

	bestDelimiter := ","
	maxConsistency := 0

	for _, delimiter := range delimiters {
		consistency := s.calculateDelimiterConsistency(testLines, delimiter)
		if consistency > maxConsistency {
			maxConsistency = consistency
			bestDelimiter = delimiter
		}
	}

	if maxConsistency == 0 {
		return ",", fmt.Errorf("could not detect delimiter, defaulting to comma")
	}

	return bestDelimiter, nil
}

// ReadCSVStream creates a streaming CSV reader
func (s *FileService) ReadCSVStream(reader io.Reader, delimiter string, hasHeader bool) (<-chan []string, <-chan error) {
	recordChan := make(chan []string, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(recordChan)
		defer close(errorChan)

		csvReader := csv.NewReader(reader)
		csvReader.Comma = rune(delimiter[0])
		csvReader.LazyQuotes = true
		csvReader.TrimLeadingSpace = true

		// Skip header if present
		if hasHeader {
			_, err := csvReader.Read()
			if err != nil {
				if err != io.EOF {
					errorChan <- fmt.Errorf("error reading CSV header: %w", err)
				}
				return
			}
		}

		// Read records
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errorChan <- fmt.Errorf("error reading CSV record: %w", err)
				break
			}

			// Send record to channel
			select {
			case recordChan <- record:
			default:
				// Channel full, could implement backpressure here
				errorChan <- fmt.Errorf("record channel full, processing too slow")
				return
			}
		}
	}()

	return recordChan, errorChan
}

// Helper methods

func (s *FileService) calculateDelimiterConsistency(lines []string, delimiter string) int {
	if len(lines) == 0 {
		return 0
	}

	// Count fields in each line
	fieldCounts := make([]int, 0, len(lines))
	totalFields := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Simple field counting (doesn't handle quoted fields perfectly, but good enough for detection)
		fields := strings.Split(line, delimiter)
		fieldCount := len(fields)
		fieldCounts = append(fieldCounts, fieldCount)
		totalFields += fieldCount
	}

	if len(fieldCounts) == 0 {
		return 0
	}

	// Calculate consistency (how many lines have the same number of fields as the first line)
	expectedFieldCount := fieldCounts[0]
	consistentLines := 0

	for _, count := range fieldCounts {
		if count == expectedFieldCount {
			consistentLines++
		}
	}

	// Return a score based on consistency and total field count
	// More consistent = higher score, more fields = higher score (within reason)
	consistencyScore := (consistentLines * 100) / len(fieldCounts)
	fieldCountScore := min(expectedFieldCount * 10, 50) // Cap field count bonus at 50

	return consistencyScore + fieldCountScore
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Additional utility methods for CSV processing

// ValidateCSVStructure performs basic structural validation of CSV data
func (s *FileService) ValidateCSVStructure(data []byte, delimiter string, hasHeader bool) error {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = rune(delimiter[0])
	reader.LazyQuotes = true

	var expectedFieldCount int
	lineNumber := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("CSV parsing error at line %d: %w", lineNumber+1, err)
		}

		lineNumber++

		// Skip empty records
		if len(record) == 1 && strings.TrimSpace(record[0]) == "" {
			continue
		}

		// Set expected field count from first record
		if expectedFieldCount == 0 {
			expectedFieldCount = len(record)
			if expectedFieldCount == 0 {
				return fmt.Errorf("first record is empty")
			}
		}

		// Check field count consistency
		if len(record) != expectedFieldCount {
			return fmt.Errorf("inconsistent field count at line %d: expected %d, got %d",
				lineNumber, expectedFieldCount, len(record))
		}
	}

	if lineNumber == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	if hasHeader && lineNumber == 1 {
		return fmt.Errorf("CSV file has header but no data rows")
	}

	return nil
}

// ExtractCSVSample extracts a sample of CSV data for preview purposes
func (s *FileService) ExtractCSVSample(data []byte, delimiter string, hasHeader bool, sampleSize int) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = rune(delimiter[0])
	reader.LazyQuotes = true

	var records [][]string
	count := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV sample: %w", err)
		}

		// Skip header if not needed in sample
		if count == 0 && hasHeader {
			records = append(records, record) // Include header in sample
			count++
			continue
		}

		records = append(records, record)
		count++

		if count >= sampleSize {
			break
		}
	}

	return records, nil
}

// GetCSVStats returns basic statistics about the CSV file
func (s *FileService) GetCSVStats(data []byte, delimiter string, hasHeader bool) (map[string]interface{}, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = rune(delimiter[0])
	reader.LazyQuotes = true

	stats := make(map[string]interface{})
	totalRows := 0
	dataRows := 0
	emptyRows := 0
	maxFieldCount := 0
	minFieldCount := -1
	inconsistentRows := 0
	expectedFieldCount := -1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV for stats: %w", err)
		}

		totalRows++
		fieldCount := len(record)

		// Check for empty rows
		isEmpty := true
		for _, field := range record {
			if strings.TrimSpace(field) != "" {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			emptyRows++
		} else {
			dataRows++
		}

		// Track field count statistics
		if fieldCount > maxFieldCount {
			maxFieldCount = fieldCount
		}
		if minFieldCount == -1 || fieldCount < minFieldCount {
			minFieldCount = fieldCount
		}

		// Set expected field count from first record
		if expectedFieldCount == -1 {
			expectedFieldCount = fieldCount
		} else if fieldCount != expectedFieldCount {
			inconsistentRows++
		}
	}

	// Adjust counts if header is present
	if hasHeader && totalRows > 0 {
		dataRows--
	}

	stats["total_rows"] = totalRows
	stats["data_rows"] = dataRows
	stats["empty_rows"] = emptyRows
	stats["max_field_count"] = maxFieldCount
	stats["min_field_count"] = minFieldCount
	stats["inconsistent_rows"] = inconsistentRows
	stats["expected_field_count"] = expectedFieldCount
	stats["has_header"] = hasHeader
	stats["delimiter"] = delimiter

	return stats, nil
}