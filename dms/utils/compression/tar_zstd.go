package compression

import (
	"archive/tar"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

// TarZstdCompressor implements compression using tar + zstd
type TarZstdCompressor struct {
	compressionLevel zstd.EncoderLevel
}

// CompressionResult contains the result of compression operation
type CompressionResult struct {
	CompressedPath   string  `json:"compressed_path"`
	OriginalSize     int64   `json:"original_size"`
	CompressedSize   int64   `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	Checksum         string  `json:"checksum"`
	Duration         int64   `json:"duration_ms"`
}

// DecompressionResult contains the result of decompression operation
type DecompressionResult struct {
	ExtractedPath string `json:"extracted_path"`
	FileCount     int    `json:"file_count"`
	TotalSize     int64  `json:"total_size"`
	Duration      int64  `json:"duration_ms"`
}

// NewTarZstdCompressor creates a new tar+zstd compressor
func NewTarZstdCompressor() *TarZstdCompressor {
	return &TarZstdCompressor{
		compressionLevel: zstd.SpeedDefault,
	}
}

// SetCompressionLevel sets the compression level
func (c *TarZstdCompressor) SetCompressionLevel(level int) {
	switch level {
	case 1:
		c.compressionLevel = zstd.SpeedFastest
	case 2:
		c.compressionLevel = zstd.SpeedDefault
	case 3:
		c.compressionLevel = zstd.SpeedBetterCompression
	case 4:
		c.compressionLevel = zstd.SpeedBestCompression
	default:
		c.compressionLevel = zstd.SpeedDefault
	}
}

// CompressFile compresses a single file using tar + zstd
func (c *TarZstdCompressor) CompressFile(inputPath, outputPath string) (*CompressionResult, error) {
	startTime := time.Now()

	// Get original file info
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat input file: %w", err)
	}

	originalSize := fileInfo.Size()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create zstd encoder
	encoder, err := zstd.NewWriter(outputFile, zstd.WithEncoderLevel(c.compressionLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(encoder)
	defer tarWriter.Close()

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create tar header
	header := &tar.Header{
		Name:    filepath.Base(inputPath),
		Size:    originalSize,
		Mode:    int64(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
	}

	// Write tar header
	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %w", err)
	}

	// Copy file content to tar
	hasher := sha256.New()
	multiWriter := io.MultiWriter(tarWriter, hasher)

	if _, err := io.Copy(multiWriter, inputFile); err != nil {
		return nil, fmt.Errorf("failed to write file to tar: %w", err)
	}

	// Close tar writer to flush
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Close encoder to flush
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zstd encoder: %w", err)
	}

	// Get compressed file size
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}

	compressedSize := compressedInfo.Size()
	compressionRatio := float64(compressedSize) / float64(originalSize)
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	duration := time.Since(startTime).Milliseconds()

	return &CompressionResult{
		CompressedPath:   outputPath,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		Checksum:         checksum,
		Duration:         duration,
	}, nil
}

// CompressDirectory compresses a directory using tar + zstd
func (c *TarZstdCompressor) CompressDirectory(inputDir, outputPath string) (*CompressionResult, error) {
	startTime := time.Now()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create zstd encoder
	encoder, err := zstd.NewWriter(outputFile, zstd.WithEncoderLevel(c.compressionLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(encoder)
	defer tarWriter.Close()

	var totalSize int64
	hasher := sha256.New()

	// Walk through directory
	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a regular file, copy content
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			multiWriter := io.MultiWriter(tarWriter, hasher)
			written, err := io.Copy(multiWriter, file)
			if err != nil {
				return err
			}
			totalSize += written
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Close writers
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zstd encoder: %w", err)
	}

	// Get compressed file size
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}

	compressedSize := compressedInfo.Size()
	compressionRatio := float64(compressedSize) / float64(totalSize)
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	duration := time.Since(startTime).Milliseconds()

	return &CompressionResult{
		CompressedPath:   outputPath,
		OriginalSize:     totalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		Checksum:         checksum,
		Duration:         duration,
	}, nil
}

// DecompressFile decompresses a tar+zstd file
func (c *TarZstdCompressor) DecompressFile(compressedPath, outputDir string) (*DecompressionResult, error) {
	startTime := time.Now()

	// Open compressed file
	compressedFile, err := os.Open(compressedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressedFile.Close()

	// Create zstd decoder
	decoder, err := zstd.NewReader(compressedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	// Create tar reader
	tarReader := tar.NewReader(decoder)

	var fileCount int
	var totalSize int64

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Create full output path
		outputPath := filepath.Join(outputDir, header.Name)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Handle different file types
		switch header.Typeflag {
		case tar.TypeReg:
			// Regular file
			file, err := os.Create(outputPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}

			written, err := io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to extract file: %w", err)
			}

			// Set file permissions and modification time
			if err := os.Chmod(outputPath, os.FileMode(header.Mode)); err != nil {
				return nil, fmt.Errorf("failed to set file mode: %w", err)
			}

			if err := os.Chtimes(outputPath, time.Now(), header.ModTime); err != nil {
				return nil, fmt.Errorf("failed to set file times: %w", err)
			}

			totalSize += written
			fileCount++

		case tar.TypeDir:
			// Directory
			if err := os.MkdirAll(outputPath, os.FileMode(header.Mode)); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	duration := time.Since(startTime).Milliseconds()

	return &DecompressionResult{
		ExtractedPath: outputDir,
		FileCount:     fileCount,
		TotalSize:     totalSize,
		Duration:      duration,
	}, nil
}

// ValidateIntegrity validates the integrity of a compressed file
func (c *TarZstdCompressor) ValidateIntegrity(compressedPath string, expectedChecksum string) error {
	// Open compressed file
	compressedFile, err := os.Open(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressedFile.Close()

	// Create zstd decoder
	decoder, err := zstd.NewReader(compressedFile)
	if err != nil {
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	// Create tar reader
	tarReader := tar.NewReader(decoder)

	hasher := sha256.New()

	// Read all content to calculate checksum
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			if _, err := io.Copy(hasher, tarReader); err != nil {
				return fmt.Errorf("failed to read file content: %w", err)
			}
		}
	}

	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// GetCompressionInfo returns information about a compressed file
func (c *TarZstdCompressor) GetCompressionInfo(compressedPath string) (map[string]interface{}, error) {
	// Get file info
	fileInfo, err := os.Stat(compressedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}

	// Open and peek into the file
	compressedFile, err := os.Open(compressedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressedFile.Close()

	// Create zstd decoder
	decoder, err := zstd.NewReader(compressedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	// Create tar reader
	tarReader := tar.NewReader(decoder)

	var fileCount int
	var totalOriginalSize int64

	// Count files and calculate original size
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			fileCount++
			totalOriginalSize += header.Size
		}
	}

	compressionRatio := float64(fileInfo.Size()) / float64(totalOriginalSize)

	return map[string]interface{}{
		"compressed_size":    fileInfo.Size(),
		"original_size":      totalOriginalSize,
		"compression_ratio":  compressionRatio,
		"file_count":         fileCount,
		"compression_type":   "tar+zstd",
		"modified_time":      fileInfo.ModTime(),
	}, nil
}