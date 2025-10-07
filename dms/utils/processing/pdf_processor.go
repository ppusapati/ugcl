package processing

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"go.uber.org/zap"
)

// PDFProcessor handles PDF document processing
type PDFProcessor struct {
	logger    *zap.Logger
	tempDir   string
	maxMemory int64 // Maximum memory usage in bytes
}

// PDFInfo contains information about a PDF document
type PDFInfo struct {
	PageCount    int                    `json:"page_count"`
	Title        string                 `json:"title"`
	Author       string                 `json:"author"`
	Subject      string                 `json:"subject"`
	Creator      string                 `json:"creator"`
	Producer     string                 `json:"producer"`
	Keywords     string                 `json:"keywords"`
	CreationDate string                 `json:"creation_date"`
	ModDate      string                 `json:"mod_date"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThumbnailOptions defines options for thumbnail generation
type ThumbnailOptions struct {
	Page      int     `json:"page"`       // Page number (0-based)
	Width     int     `json:"width"`      // Target width
	Height    int     `json:"height"`     // Target height
	Scale     float64 `json:"scale"`      // Scale factor
	Format    string  `json:"format"`     // Output format: "png", "jpg"
	Quality   int     `json:"quality"`    // JPEG quality (1-100)
	KeepRatio bool    `json:"keep_ratio"` // Maintain aspect ratio
}

// PageRenderOptions defines options for page rendering
type PageRenderOptions struct {
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Scale     float64 `json:"scale"`
	Rotation  int     `json:"rotation"`  // Rotation in degrees
	Format    string  `json:"format"`    // "png", "jpg", "svg"
	Quality   int     `json:"quality"`   // JPEG quality
	DPI       int     `json:"dpi"`       // Dots per inch
}

// TextExtractionResult contains extracted text information
type TextExtractionResult struct {
	Text      string            `json:"text"`
	PageTexts map[int]string    `json:"page_texts"` // Text by page number
	Metadata  map[string]string `json:"metadata"`
}

// NewPDFProcessor creates a new PDF processor
func NewPDFProcessor(logger *zap.Logger, tempDir string) *PDFProcessor {
	return &PDFProcessor{
		logger:    logger,
		tempDir:   tempDir,
		maxMemory: 512 * 1024 * 1024, // 512MB default
	}
}

// GetPDFInfo extracts metadata and information from a PDF
func (p *PDFProcessor) GetPDFInfo(ctx context.Context, filePath string) (*PDFInfo, error) {
	// Open PDF document
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Get basic info
	info := &PDFInfo{
		PageCount: doc.NumPage(),
		Metadata:  make(map[string]interface{}),
	}

	// Extract metadata
	metadata := doc.Metadata()

	// Parse standard metadata fields
	if title, ok := metadata["title"]; ok {
		info.Title = title
	}
	if author, ok := metadata["author"]; ok {
		info.Author = author
	}
	if subject, ok := metadata["subject"]; ok {
		info.Subject = subject
	}
	if creator, ok := metadata["creator"]; ok {
		info.Creator = creator
	}
	if producer, ok := metadata["producer"]; ok {
		info.Producer = producer
	}
	if keywords, ok := metadata["keywords"]; ok {
		info.Keywords = keywords
	}
	if creationDate, ok := metadata["creationDate"]; ok {
		info.CreationDate = creationDate
	}
	if modDate, ok := metadata["modDate"]; ok {
		info.ModDate = modDate
	}

	// Store all metadata
	info.Metadata = metadata

	p.logger.Debug("Extracted PDF info",
		zap.String("file", filePath),
		zap.Int("pages", info.PageCount),
		zap.String("title", info.Title))

	return info, nil
}

// GenerateThumbnail generates a thumbnail for a specific page
func (p *PDFProcessor) GenerateThumbnail(ctx context.Context, filePath string, opts ThumbnailOptions) ([]byte, error) {
	// Validate options
	if opts.Format == "" {
		opts.Format = "png"
	}
	if opts.Quality == 0 {
		opts.Quality = 85
	}
	if opts.Scale == 0 {
		opts.Scale = 1.0
	}

	// Open PDF document
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Validate page number
	if opts.Page < 0 || opts.Page >= doc.NumPage() {
		return nil, fmt.Errorf("page %d is out of range (0-%d)", opts.Page, doc.NumPage()-1)
	}

	// Render page as image
	img, err := doc.Image(opts.Page)
	if err != nil {
		return nil, fmt.Errorf("failed to render page %d: %w", opts.Page, err)
	}

	// Resize image if needed
	if opts.Width > 0 || opts.Height > 0 {
		img = p.resizeImage(img, opts.Width, opts.Height, opts.KeepRatio)
	} else if opts.Scale != 1.0 {
		bounds := img.Bounds()
		newWidth := int(float64(bounds.Dx()) * opts.Scale)
		newHeight := int(float64(bounds.Dy()) * opts.Scale)
		img = p.resizeImage(img, newWidth, newHeight, true)
	}

	// Encode image to bytes
	return p.encodeImage(img, opts.Format, opts.Quality)
}

// RenderPage renders a PDF page to an image
func (p *PDFProcessor) RenderPage(ctx context.Context, filePath string, pageNum int, opts PageRenderOptions) ([]byte, error) {
	// Validate options
	if opts.Format == "" {
		opts.Format = "png"
	}
	if opts.Quality == 0 {
		opts.Quality = 90
	}
	if opts.Scale == 0 {
		opts.Scale = 1.0
	}
	if opts.DPI == 0 {
		opts.DPI = 150
	}

	// Open PDF document
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Validate page number
	if pageNum < 0 || pageNum >= doc.NumPage() {
		return nil, fmt.Errorf("page %d is out of range (0-%d)", pageNum, doc.NumPage()-1)
	}

	// Render page with options
	img, err := doc.ImageDPI(pageNum, float64(opts.DPI))
	if err != nil {
		return nil, fmt.Errorf("failed to render page %d: %w", pageNum, err)
	}

	// Apply transformations
	if opts.Width > 0 || opts.Height > 0 {
		img = p.resizeImage(img, opts.Width, opts.Height, true)
	} else if opts.Scale != 1.0 {
		bounds := img.Bounds()
		newWidth := int(float64(bounds.Dx()) * opts.Scale)
		newHeight := int(float64(bounds.Dy()) * opts.Scale)
		img = p.resizeImage(img, newWidth, newHeight, true)
	}

	// Apply rotation if needed
	if opts.Rotation != 0 {
		img = p.rotateImage(img, opts.Rotation)
	}

	// Encode image
	return p.encodeImage(img, opts.Format, opts.Quality)
}

// ExtractText extracts text from PDF pages
func (p *PDFProcessor) ExtractText(ctx context.Context, filePath string, pageNum int) (*TextExtractionResult, error) {
	// Open PDF document
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	result := &TextExtractionResult{
		PageTexts: make(map[int]string),
		Metadata:  make(map[string]string),
	}

	// Extract text from specific page or all pages
	if pageNum >= 0 {
		// Single page
		if pageNum >= doc.NumPage() {
			return nil, fmt.Errorf("page %d is out of range (0-%d)", pageNum, doc.NumPage()-1)
		}

		text, err := doc.Text(pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
		}

		result.Text = text
		result.PageTexts[pageNum] = text
	} else {
		// All pages
		var allText strings.Builder
		for i := 0; i < doc.NumPage(); i++ {
			text, err := doc.Text(i)
			if err != nil {
				p.logger.Warn("Failed to extract text from page",
					zap.Int("page", i),
					zap.Error(err))
				continue
			}

			result.PageTexts[i] = text
			allText.WriteString(text)
			if i < doc.NumPage()-1 {
				allText.WriteString("\n\n")
			}
		}
		result.Text = allText.String()
	}

	// Add metadata
	result.Metadata["page_count"] = fmt.Sprintf("%d", doc.NumPage())
	result.Metadata["extraction_method"] = "fitz"

	p.logger.Debug("Extracted text from PDF",
		zap.String("file", filePath),
		zap.Int("page", pageNum),
		zap.Int("text_length", len(result.Text)))

	return result, nil
}

// ConvertToImages converts all PDF pages to images
func (p *PDFProcessor) ConvertToImages(ctx context.Context, filePath, outputDir string, opts PageRenderOptions) ([]string, error) {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open PDF document
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Set default options
	if opts.Format == "" {
		opts.Format = "png"
	}
	if opts.Quality == 0 {
		opts.Quality = 90
	}
	if opts.DPI == 0 {
		opts.DPI = 150
	}

	var imagePaths []string

	// Convert each page
	for i := 0; i < doc.NumPage(); i++ {
		// Render page
		img, err := doc.ImageDPI(i, float64(opts.DPI))
		if err != nil {
			p.logger.Warn("Failed to render page",
				zap.Int("page", i),
				zap.Error(err))
			continue
		}

		// Apply transformations
		if opts.Width > 0 || opts.Height > 0 {
			img = p.resizeImage(img, opts.Width, opts.Height, true)
		}

		if opts.Rotation != 0 {
			img = p.rotateImage(img, opts.Rotation)
		}

		// Save image
		filename := fmt.Sprintf("page_%03d.%s", i+1, opts.Format)
		imagePath := filepath.Join(outputDir, filename)

		if err := p.saveImage(img, imagePath, opts.Format, opts.Quality); err != nil {
			p.logger.Warn("Failed to save page image",
				zap.Int("page", i),
				zap.String("path", imagePath),
				zap.Error(err))
			continue
		}

		imagePaths = append(imagePaths, imagePath)
	}

	p.logger.Info("Converted PDF to images",
		zap.String("file", filePath),
		zap.Int("pages", len(imagePaths)),
		zap.String("output_dir", outputDir))

	return imagePaths, nil
}

// ValidatePDF checks if a file is a valid PDF
func (p *PDFProcessor) ValidatePDF(ctx context.Context, filePath string) error {
	doc, err := fitz.New(filePath)
	if err != nil {
		return fmt.Errorf("invalid PDF file: %w", err)
	}
	defer doc.Close()

	// Basic validation - check if we can get page count
	if doc.NumPage() <= 0 {
		return fmt.Errorf("PDF has no pages")
	}

	return nil
}

// Private helper methods

func (p *PDFProcessor) resizeImage(img image.Image, width, height int, keepRatio bool) image.Image {
	// This is a placeholder implementation
	// In a real application, you would use a proper image resizing library
	// like "github.com/disintegration/imaging"
	return img
}

func (p *PDFProcessor) rotateImage(img image.Image, degrees int) image.Image {
	// This is a placeholder implementation
	// In a real application, you would implement proper image rotation
	return img
}

func (p *PDFProcessor) encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp(p.tempDir, "image_*."+format)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Encode based on format
	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(tempFile, img)
	case "jpg", "jpeg":
		err = jpeg.Encode(tempFile, img, &jpeg.Options{Quality: quality})
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Read back the encoded data
	tempFile.Seek(0, 0)
	return io.ReadAll(tempFile)
}

func (p *PDFProcessor) saveImage(img image.Image, path, format string, quality int) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	switch strings.ToLower(format) {
	case "png":
		return png.Encode(file, img)
	case "jpg", "jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// SetMaxMemory sets the maximum memory usage
func (p *PDFProcessor) SetMaxMemory(bytes int64) {
	p.maxMemory = bytes
}

// GetSupportedFormats returns supported output formats
func (p *PDFProcessor) GetSupportedFormats() []string {
	return []string{"png", "jpg", "jpeg"}
}