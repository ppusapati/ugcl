package processing

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"go.uber.org/zap"
)

// OCRProcessor handles Optical Character Recognition
type OCRProcessor struct {
	logger       *zap.Logger
	tempDir      string
	tesseractPath string
	languages    []string
	confidence   float64
}

// OCRResult contains OCR extraction results
type OCRResult struct {
	Text       string            `json:"text"`
	Confidence float64           `json:"confidence"`
	Language   string            `json:"language"`
	WordCount  int               `json:"word_count"`
	PageTexts  map[int]string    `json:"page_texts,omitempty"`
	BoundingBoxes []BoundingBox  `json:"bounding_boxes,omitempty"`
	Metadata   map[string]string `json:"metadata"`
}

// BoundingBox represents text position information
type BoundingBox struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Page       int     `json:"page"`
}

// OCROptions defines OCR processing options
type OCROptions struct {
	Language           string   `json:"language"`            // Language code (e.g., "eng", "fra")
	Languages          []string `json:"languages"`           // Multiple languages
	PSM                int      `json:"psm"`                 // Page Segmentation Mode
	OEM                int      `json:"oem"`                 // OCR Engine Mode
	MinConfidence      float64  `json:"min_confidence"`      // Minimum confidence threshold
	EnableBoundingBoxes bool    `json:"enable_bounding_boxes"` // Extract bounding box information
	PreprocessImage    bool     `json:"preprocess_image"`    // Apply image preprocessing
	DPI                int      `json:"dpi"`                 // Image DPI
	ConfigVars         map[string]string `json:"config_vars"` // Tesseract config variables
}

// NewOCRProcessor creates a new OCR processor
func NewOCRProcessor(logger *zap.Logger, tempDir string) *OCRProcessor {
	return &OCRProcessor{
		logger:     logger,
		tempDir:    tempDir,
		languages:  []string{"eng"}, // Default to English
		confidence: 0.0,             // No confidence threshold by default
	}
}

// ProcessImage performs OCR on an image
func (o *OCRProcessor) ProcessImage(ctx context.Context, imagePath string, opts OCROptions) (*OCRResult, error) {
	// Create Tesseract client
	client := gosseract.NewClient()
	defer client.Close()

	// Set language
	language := opts.Language
	if language == "" && len(opts.Languages) > 0 {
		language = strings.Join(opts.Languages, "+")
	}
	if language == "" {
		language = "eng"
	}

	if err := client.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("failed to set OCR language %s: %w", language, err)
	}

	// Set image source
	if err := client.SetImage(imagePath); err != nil {
		return nil, fmt.Errorf("failed to set image source: %w", err)
	}

	// Configure Tesseract options
	if opts.PSM > 0 {
		client.SetPageSegMode(gosseract.PageSegMode(opts.PSM))
	} else {
		client.SetPageSegMode(gosseract.PSM_AUTO) // Default to auto
	}

	if opts.OEM > 0 {
		// Set OCR Engine Mode if specified
		client.SetVariable("tessedit_ocr_engine_mode", fmt.Sprintf("%d", opts.OEM))
	}

	// Set additional config variables
	for key, value := range opts.ConfigVars {
		client.SetVariable(key, value)
	}

	// Set DPI if specified
	if opts.DPI > 0 {
		client.SetVariable("user_defined_dpi", fmt.Sprintf("%d", opts.DPI))
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Get confidence
	confidence := float64(client.GetAvailableLanguages().GetConfidence())

	// Prepare result
	result := &OCRResult{
		Text:       strings.TrimSpace(text),
		Confidence: confidence,
		Language:   language,
		WordCount:  len(strings.Fields(text)),
		Metadata:   make(map[string]string),
	}

	// Add metadata
	result.Metadata["tesseract_version"] = gosseract.Version()
	result.Metadata["language"] = language
	result.Metadata["psm"] = fmt.Sprintf("%d", opts.PSM)
	result.Metadata["input_file"] = imagePath

	// Extract bounding boxes if requested
	if opts.EnableBoundingBoxes {
		boundingBoxes, err := o.extractBoundingBoxes(client)
		if err != nil {
			o.logger.Warn("Failed to extract bounding boxes", zap.Error(err))
		} else {
			result.BoundingBoxes = boundingBoxes
		}
	}

	o.logger.Debug("OCR processing completed",
		zap.String("image", imagePath),
		zap.String("language", language),
		zap.Float64("confidence", confidence),
		zap.Int("text_length", len(text)))

	return result, nil
}

// ProcessPDFPages performs OCR on PDF pages (assumes images have been extracted)
func (o *OCRProcessor) ProcessPDFPages(ctx context.Context, imagePaths []string, opts OCROptions) (*OCRResult, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no images provided for OCR")
	}

	var allText strings.Builder
	var totalConfidence float64
	var wordCount int
	pageTexts := make(map[int]string)
	var allBoundingBoxes []BoundingBox

	// Process each page
	for i, imagePath := range imagePaths {
		pageResult, err := o.ProcessImage(ctx, imagePath, opts)
		if err != nil {
			o.logger.Warn("Failed to process page",
				zap.Int("page", i),
				zap.String("image", imagePath),
				zap.Error(err))
			continue
		}

		// Accumulate results
		pageTexts[i] = pageResult.Text
		allText.WriteString(pageResult.Text)
		if i < len(imagePaths)-1 {
			allText.WriteString("\n\n")
		}

		totalConfidence += pageResult.Confidence
		wordCount += pageResult.WordCount

		// Add page number to bounding boxes
		for _, bbox := range pageResult.BoundingBoxes {
			bbox.Page = i
			allBoundingBoxes = append(allBoundingBoxes, bbox)
		}
	}

	// Calculate average confidence
	avgConfidence := totalConfidence / float64(len(imagePaths))

	// Create combined result
	result := &OCRResult{
		Text:          allText.String(),
		Confidence:    avgConfidence,
		Language:      opts.Language,
		WordCount:     wordCount,
		PageTexts:     pageTexts,
		BoundingBoxes: allBoundingBoxes,
		Metadata:      make(map[string]string),
	}

	// Add metadata
	result.Metadata["pages_processed"] = fmt.Sprintf("%d", len(imagePaths))
	result.Metadata["average_confidence"] = fmt.Sprintf("%.2f", avgConfidence)

	o.logger.Info("OCR processing completed for PDF",
		zap.Int("pages", len(imagePaths)),
		zap.Float64("avg_confidence", avgConfidence),
		zap.Int("total_words", wordCount))

	return result, nil
}

// ProcessDocument performs OCR on various document types
func (o *OCRProcessor) ProcessDocument(ctx context.Context, filePath string, opts OCROptions) (*OCRResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return o.processPDFDocument(ctx, filePath, opts)
	case ".png", ".jpg", ".jpeg", ".tiff", ".bmp":
		return o.ProcessImage(ctx, filePath, opts)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// PreprocessImage applies image preprocessing to improve OCR accuracy
func (o *OCRProcessor) PreprocessImage(ctx context.Context, inputPath, outputPath string) error {
	// This is a placeholder for image preprocessing
	// In a real implementation, you would apply various image processing techniques:
	// - Noise reduction
	// - Contrast enhancement
	// - Deskewing
	// - Binarization
	// - Resolution enhancement

	// For now, just copy the file
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input image: %w", err)
	}
	defer input.Close()

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output image: %w", err)
	}
	defer output.Close()

	_, err = output.ReadFrom(input)
	if err != nil {
		return fmt.Errorf("failed to copy image: %w", err)
	}

	o.logger.Debug("Image preprocessing completed",
		zap.String("input", inputPath),
		zap.String("output", outputPath))

	return nil
}

// GetSupportedLanguages returns available OCR languages
func (o *OCRProcessor) GetSupportedLanguages() ([]string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Get available languages
	languages := client.GetAvailableLanguages()
	if languages == nil {
		return nil, fmt.Errorf("failed to get available languages")
	}

	return strings.Split(languages.String(), "\n"), nil
}

// ValidateLanguage checks if a language is supported
func (o *OCRProcessor) ValidateLanguage(language string) error {
	supportedLangs, err := o.GetSupportedLanguages()
	if err != nil {
		return fmt.Errorf("failed to get supported languages: %w", err)
	}

	for _, lang := range supportedLangs {
		if strings.TrimSpace(lang) == language {
			return nil
		}
	}

	return fmt.Errorf("language %s is not supported", language)
}

// Private methods

func (o *OCRProcessor) processPDFDocument(ctx context.Context, filePath string, opts OCROptions) (*OCRResult, error) {
	// First, we need to convert PDF pages to images
	// This requires the PDF processor
	pdfProcessor := NewPDFProcessor(o.logger, o.tempDir)

	// Create temporary directory for images
	tempImageDir := filepath.Join(o.tempDir, fmt.Sprintf("ocr_images_%d", os.Getpid()))
	if err := os.MkdirAll(tempImageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp image directory: %w", err)
	}
	defer os.RemoveAll(tempImageDir)

	// Convert PDF to images
	renderOpts := PageRenderOptions{
		Format:  "png",
		DPI:     300, // High DPI for better OCR accuracy
		Quality: 100,
	}

	imagePaths, err := pdfProcessor.ConvertToImages(ctx, filePath, tempImageDir, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// Preprocess images if requested
	if opts.PreprocessImage {
		for i, imagePath := range imagePaths {
			preprocessedPath := filepath.Join(tempImageDir, fmt.Sprintf("preprocessed_%d.png", i))
			if err := o.PreprocessImage(ctx, imagePath, preprocessedPath); err != nil {
				o.logger.Warn("Failed to preprocess image",
					zap.String("image", imagePath),
					zap.Error(err))
			} else {
				imagePaths[i] = preprocessedPath
			}
		}
	}

	// Perform OCR on all pages
	return o.ProcessPDFPages(ctx, imagePaths, opts)
}

func (o *OCRProcessor) extractBoundingBoxes(client *gosseract.Client) ([]BoundingBox, error) {
	// This is a placeholder for bounding box extraction
	// In a real implementation, you would use Tesseract's API to get
	// word-level or character-level bounding box information

	var boundingBoxes []BoundingBox

	// Placeholder implementation
	// You would typically use client.GetBoundingBoxes() or similar
	// depending on the gosseract library version

	return boundingBoxes, nil
}

// SetDefaultLanguages sets the default OCR languages
func (o *OCRProcessor) SetDefaultLanguages(languages []string) {
	o.languages = languages
}

// SetMinConfidence sets the minimum confidence threshold
func (o *OCRProcessor) SetMinConfidence(confidence float64) {
	o.confidence = confidence
}

// IsTextDetected checks if meaningful text was detected
func (o *OCRProcessor) IsTextDetected(result *OCRResult, minWords int) bool {
	if result == nil {
		return false
	}

	// Check word count
	if result.WordCount < minWords {
		return false
	}

	// Check confidence if set
	if o.confidence > 0 && result.Confidence < o.confidence {
		return false
	}

	// Check for meaningful text (not just noise)
	cleanText := strings.TrimSpace(result.Text)
	if len(cleanText) < 10 {
		return false
	}

	return true
}