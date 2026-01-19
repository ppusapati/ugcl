package watermark

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	"p9e.in/ugcl/dms/models"
)

// WatermarkGenerator handles dynamic watermark generation
type WatermarkGenerator struct {
	tempDir       string
	defaultFont   string
	logoCache     map[string]image.Image
}

// WatermarkData contains dynamic data for watermark generation
type WatermarkData struct {
	Username     string    `json:"username"`
	Timestamp    time.Time `json:"timestamp"`
	DocumentName string    `json:"document_name"`
	IPAddress    string    `json:"ip_address"`
	CompanyName  string    `json:"company_name"`
	Department   string    `json:"department"`
	AccessLevel  string    `json:"access_level"`
	SessionID    string    `json:"session_id"`
}

// WatermarkPosition represents watermark positioning
type WatermarkPosition struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Rotation float64 `json:"rotation"`
}

// GeneratedWatermark contains the result of watermark generation
type GeneratedWatermark struct {
	ImagePath    string                 `json:"image_path"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	Position     WatermarkPosition      `json:"position"`
	Metadata     map[string]interface{} `json:"metadata"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

// NewWatermarkGenerator creates a new watermark generator
func NewWatermarkGenerator(tempDir string) *WatermarkGenerator {
	return &WatermarkGenerator{
		tempDir:     tempDir,
		defaultFont: "Arial",
		logoCache:   make(map[string]image.Image),
	}
}

// GenerateWatermark generates a watermark based on configuration and dynamic data
func (g *WatermarkGenerator) GenerateWatermark(config *models.WatermarkConfig, data *WatermarkData) (*GeneratedWatermark, error) {
	switch config.Type {
	case "text":
		return g.generateTextWatermark(config, data)
	case "image":
		return g.generateImageWatermark(config, data)
	case "combined":
		return g.generateCombinedWatermark(config, data)
	default:
		return nil, fmt.Errorf("unsupported watermark type: %s", config.Type)
	}
}

// generateTextWatermark creates a text-based watermark
func (g *WatermarkGenerator) generateTextWatermark(config *models.WatermarkConfig, data *WatermarkData) (*GeneratedWatermark, error) {
	// Parse text template
	text, err := g.parseTextTemplate(config.TextTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %w", err)
	}

	// Create image for text rendering
	img := image.NewRGBA(image.Rect(0, 0, config.MaxWidth, config.MaxHeight))

	// Fill with transparent background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	// Parse color
	textColor, err := g.parseColor(config.FontColor)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font color: %w", err)
	}

	// Create font configuration
	fontConfig := FontConfig{
		Family: config.FontFamily,
		Size:   config.FontSize,
		Bold:   config.FontBold,
		Italic: config.FontItalic,
		Color:  textColor,
	}

	// Render text
	textImg, err := g.renderText(text, fontConfig, config.MaxWidth, config.MaxHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to render text: %w", err)
	}

	// Apply opacity
	if config.Opacity < 1.0 {
		textImg = g.applyOpacity(textImg, config.Opacity)
	}

	// Apply shadow if enabled
	if config.Shadow {
		shadowColor, _ := g.parseColor(config.ShadowColor)
		textImg = g.applyShadow(textImg, shadowColor, config.ShadowOffsetX, config.ShadowOffsetY)
	}

	// Save to temp file
	outputPath := g.getTempFilePath("watermark", "png")
	if err := g.saveImageToPNG(textImg, outputPath); err != nil {
		return nil, fmt.Errorf("failed to save watermark image: %w", err)
	}

	// Calculate position
	position := g.calculatePosition(config.Position, textImg.Bounds().Dx(), textImg.Bounds().Dy(), config.OffsetX, config.OffsetY)
	position.Rotation = config.Rotation

	return &GeneratedWatermark{
		ImagePath: outputPath,
		Width:     textImg.Bounds().Dx(),
		Height:    textImg.Bounds().Dy(),
		Position:  position,
		Metadata: map[string]interface{}{
			"type":      "text",
			"text":      text,
			"font":      fontConfig,
			"opacity":   config.Opacity,
			"config_id": config.ID,
		},
		GeneratedAt: time.Now(),
	}, nil
}

// generateImageWatermark creates an image-based watermark
func (g *WatermarkGenerator) generateImageWatermark(config *models.WatermarkConfig, data *WatermarkData) (*GeneratedWatermark, error) {
	var watermarkImg image.Image
	var err error

	// Load base image
	if config.ShowCompanyLogo && config.LogoPath != "" {
		watermarkImg, err = g.loadImage(config.LogoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load logo: %w", err)
		}
	} else if config.ImagePath != "" {
		watermarkImg, err = g.loadImage(config.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load watermark image: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no image path specified for image watermark")
	}

	// Resize if necessary
	if config.MaxWidth > 0 || config.MaxHeight > 0 {
		watermarkImg = imaging.Fit(watermarkImg, config.MaxWidth, config.MaxHeight, imaging.Lanczos)
	}

	// Apply scale
	if config.Scale != 1.0 {
		newWidth := int(float64(watermarkImg.Bounds().Dx()) * config.Scale)
		newHeight := int(float64(watermarkImg.Bounds().Dy()) * config.Scale)
		watermarkImg = imaging.Resize(watermarkImg, newWidth, newHeight, imaging.Lanczos)
	}

	// Apply opacity
	if config.Opacity < 1.0 {
		watermarkImg = g.applyOpacity(watermarkImg, config.Opacity)
	}

	// Save to temp file
	outputPath := g.getTempFilePath("watermark", "png")
	if err := g.saveImageToPNG(watermarkImg, outputPath); err != nil {
		return nil, fmt.Errorf("failed to save watermark image: %w", err)
	}

	// Calculate position
	position := g.calculatePosition(config.Position, watermarkImg.Bounds().Dx(), watermarkImg.Bounds().Dy(), config.OffsetX, config.OffsetY)
	position.Rotation = config.Rotation

	return &GeneratedWatermark{
		ImagePath: outputPath,
		Width:     watermarkImg.Bounds().Dx(),
		Height:    watermarkImg.Bounds().Dy(),
		Position:  position,
		Metadata: map[string]interface{}{
			"type":      "image",
			"opacity":   config.Opacity,
			"scale":     config.Scale,
			"config_id": config.ID,
		},
		GeneratedAt: time.Now(),
	}, nil
}

// generateCombinedWatermark creates a combined text and image watermark
func (g *WatermarkGenerator) generateCombinedWatermark(config *models.WatermarkConfig, data *WatermarkData) (*GeneratedWatermark, error) {
	// Generate text watermark
	textWatermark, err := g.generateTextWatermark(config, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text watermark: %w", err)
	}

	// Load text image
	textImg, err := g.loadImage(textWatermark.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load text watermark: %w", err)
	}

	var combinedImg image.Image

	// Load and combine with logo if specified
	if config.ShowCompanyLogo && config.LogoPath != "" {
		logoImg, err := g.loadImage(config.LogoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load logo: %w", err)
		}

		// Resize logo to fit with text
		logoImg = imaging.Fit(logoImg, config.MaxWidth/3, config.MaxHeight/2, imaging.Lanczos)

		// Combine text and logo
		combinedImg = g.combineImages(logoImg, textImg, "horizontal")
	} else {
		combinedImg = textImg
	}

	// Apply final opacity
	if config.Opacity < 1.0 {
		combinedImg = g.applyOpacity(combinedImg, config.Opacity)
	}

	// Save combined image
	outputPath := g.getTempFilePath("watermark", "png")
	if err := g.saveImageToPNG(combinedImg, outputPath); err != nil {
		return nil, fmt.Errorf("failed to save combined watermark: %w", err)
	}

	// Calculate position
	position := g.calculatePosition(config.Position, combinedImg.Bounds().Dx(), combinedImg.Bounds().Dy(), config.OffsetX, config.OffsetY)
	position.Rotation = config.Rotation

	// Clean up temporary text watermark
	os.Remove(textWatermark.ImagePath)

	return &GeneratedWatermark{
		ImagePath: outputPath,
		Width:     combinedImg.Bounds().Dx(),
		Height:    combinedImg.Bounds().Dy(),
		Position:  position,
		Metadata: map[string]interface{}{
			"type":      "combined",
			"opacity":   config.Opacity,
			"config_id": config.ID,
		},
		GeneratedAt: time.Now(),
	}, nil
}

// parseTextTemplate parses a text template with dynamic data
func (g *WatermarkGenerator) parseTextTemplate(templateStr string, data *WatermarkData) (string, error) {
	tmpl, err := template.New("watermark").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// FontConfig represents font configuration
type FontConfig struct {
	Family string
	Size   int
	Bold   bool
	Italic bool
	Color  color.RGBA
}

// renderText renders text with specified font configuration
func (g *WatermarkGenerator) renderText(text string, config FontConfig, maxWidth, maxHeight int) (image.Image, error) {
	// This is a simplified text rendering implementation
	// In a real application, you would use a proper font rendering library like freetype-go

	// Create a basic text image
	img := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))

	// Fill with transparent background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	// For now, create a simple colored rectangle as placeholder
	// In real implementation, render actual text using font libraries
	textBounds := image.Rect(10, 10, maxWidth-10, config.Size+20)
	draw.Draw(img, textBounds, &image.Uniform{config.Color}, image.Point{}, draw.Src)

	return img, nil
}

// Helper functions

func (g *WatermarkGenerator) parseColor(colorStr string) (color.RGBA, error) {
	// Simple hex color parsing
	if len(colorStr) == 7 && colorStr[0] == '#' {
		var r, g, b uint8
		if _, err := fmt.Sscanf(colorStr, "#%02x%02x%02x", &r, &g, &b); err != nil {
			return color.RGBA{}, err
		}
		return color.RGBA{r, g, b, 255}, nil
	}
	return color.RGBA{0, 0, 0, 255}, nil // Default to black
}

func (g *WatermarkGenerator) loadImage(path string) (image.Image, error) {
	if img, exists := g.logoCache[path]; exists {
		return img, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	g.logoCache[path] = img
	return img, nil
}

func (g *WatermarkGenerator) applyOpacity(img image.Image, opacity float64) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			newA := uint8(float64(a>>8) * opacity)
			newImg.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), newA})
		}
	}

	return newImg
}

func (g *WatermarkGenerator) applyShadow(img image.Image, shadowColor color.RGBA, offsetX, offsetY int) image.Image {
	bounds := img.Bounds()
	newBounds := image.Rect(0, 0, bounds.Dx()+offsetX, bounds.Dy()+offsetY)
	newImg := image.NewRGBA(newBounds)

	// Draw shadow
	draw.DrawMask(newImg, bounds.Add(image.Pt(offsetX, offsetY)), &image.Uniform{shadowColor}, image.Point{}, img, bounds.Min, draw.Over)

	// Draw original image
	draw.DrawMask(newImg, bounds, img, bounds.Min, img, bounds.Min, draw.Over)

	return newImg
}

func (g *WatermarkGenerator) combineImages(img1, img2 image.Image, direction string) image.Image {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	var newBounds image.Rectangle
	if direction == "horizontal" {
		newBounds = image.Rect(0, 0, bounds1.Dx()+bounds2.Dx(), max(bounds1.Dy(), bounds2.Dy()))
	} else {
		newBounds = image.Rect(0, 0, max(bounds1.Dx(), bounds2.Dx()), bounds1.Dy()+bounds2.Dy())
	}

	newImg := image.NewRGBA(newBounds)

	// Draw first image
	draw.Draw(newImg, bounds1, img1, bounds1.Min, draw.Src)

	// Draw second image
	var offset image.Point
	if direction == "horizontal" {
		offset = image.Pt(bounds1.Dx(), 0)
	} else {
		offset = image.Pt(0, bounds1.Dy())
	}
	draw.Draw(newImg, bounds2.Add(offset), img2, bounds2.Min, draw.Over)

	return newImg
}

func (g *WatermarkGenerator) calculatePosition(position string, width, height, offsetX, offsetY int) WatermarkPosition {
	// Default to bottom right
	x, y := offsetX, offsetY

	// This would be implemented based on document dimensions
	// For now, return relative position
	return WatermarkPosition{
		X: x,
		Y: y,
		Rotation: 0,
	}
}

func (g *WatermarkGenerator) saveImageToPNG(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func (g *WatermarkGenerator) getTempFilePath(prefix, extension string) string {
	filename := fmt.Sprintf("%s_%s.%s", prefix, uuid.New().String(), extension)
	return fmt.Sprintf("%s/%s", g.tempDir, filename)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}