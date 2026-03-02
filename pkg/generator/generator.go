package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/amsaid/gopdf-generator/pkg/elements"
	"github.com/amsaid/gopdf-generator/pkg/fonts"
	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/signintech/gopdf"
)

// PDFGenerator is the main PDF generation engine
type PDFGenerator struct {
	fontMgr *fonts.Manager
	fontDir string
	tempDir string
}

// Config holds generator configuration
type Config struct {
	FontDir    string
	TempDir    string
	EmbedFonts bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		FontDir:    "./fonts",
		TempDir:    os.TempDir(),
		EmbedFonts: true,
	}
}

// New creates a new PDF generator
func New(config *Config) (*PDFGenerator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure directories exist
	if err := os.MkdirAll(config.FontDir, 0755); err != nil {
		return nil, fmt.Errorf("creating font directory: %w", err)
	}
	if err := os.MkdirAll(config.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}

	fontMgr := fonts.NewManager(config.FontDir)
	fontMgr.RegisterStandardFonts()

	return &PDFGenerator{
		fontMgr: fontMgr,
		fontDir: config.FontDir,
		tempDir: config.TempDir,
	}, nil
}

// Generate generates a PDF from a template
func (g *PDFGenerator) Generate(template *parser.DocumentTemplate) (*bytes.Buffer, error) {
	if err := template.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Create new PDF instance for this generation
	pdf := &gopdf.GoPdf{}

	// Create a session font manager
	sessionFontMgr := g.fontMgr.Clone()

	// Initialize PDF
	if err := g.initializePDF(pdf, template); err != nil {
		return nil, fmt.Errorf("initializing PDF: %w", err)
	}

	// Register base fonts to PDF
	if err := sessionFontMgr.AddFontsToPDF(pdf); err != nil {
		return nil, fmt.Errorf("adding base fonts: %w", err)
	}

	// Register template-specific fonts
	if err := g.registerFonts(pdf, sessionFontMgr, template.Fonts); err != nil {
		return nil, fmt.Errorf("registering fonts: %w", err)
	}

	// Create element handler
	pageWidth, pageHeight := g.getPageDimensions(template)
	handler := elements.NewHandler(pdf, sessionFontMgr, pageWidth, pageHeight, template.Margin, template.DefaultFont)

	// Process elements
	for i, elem := range template.Elements {
		if err := handler.HandleElement(elem); err != nil {
			return nil, fmt.Errorf("processing element %d (%s): %w", i, elem.Type, err)
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("writing PDF: %w", err)
	}

	return &buf, nil
}

// GenerateToFile generates a PDF and saves to file
func (g *PDFGenerator) GenerateToFile(template *parser.DocumentTemplate, outputPath string) error {
	buf, err := g.Generate(template)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}

// GenerateFromJSON generates PDF from JSON data
func (g *PDFGenerator) GenerateFromJSON(data []byte) (*bytes.Buffer, error) {
	template, err := parser.ParseTemplateBytes(data)
	if err != nil {
		return nil, err
	}

	return g.Generate(template)
}

// GenerateFromReader generates PDF from JSON reader
func (g *PDFGenerator) GenerateFromReader(r io.Reader) (*bytes.Buffer, error) {
	template, err := parser.ParseTemplate(r)
	if err != nil {
		return nil, err
	}

	return g.Generate(template)
}

// initializePDF sets up the PDF document
func (g *PDFGenerator) initializePDF(pdf *gopdf.GoPdf, template *parser.DocumentTemplate) error {
	// Set page size
	pageSize := g.getPageSize(template)

	// Start PDF with page size
	pdf.Start(gopdf.Config{
		PageSize: *pageSize,
	})

	// Add first page
	pdf.AddPage()

	// Set metadata
	pdf.SetInfo(gopdf.PdfInfo{
		Title:   template.Title,
		Author:  template.Author,
		Subject: template.Subject,
		Creator: template.Creator,
	})

	return nil
}

// registerFonts registers custom fonts from template
func (g *PDFGenerator) registerFonts(pdf *gopdf.GoPdf, fontMgr *fonts.Manager, fontDefs []parser.FontDef) error {
	for _, fontDef := range fontDefs {
		if fontDef.Name == "" {
			continue
		}

		// Check if already registered
		if fontMgr.IsRegistered(fontDef.Name) {
			continue
		}

		var err error
		// If font data provided, use it
		if len(fontDef.Data) > 0 {
			err = fontMgr.RegisterFontFromBytes(fontDef.Name, fontDef.Data)
		} else if fontDef.FilePath != "" {
			err = fontMgr.RegisterFont(fontDef.Name, fontDef.FilePath)
		}

		if err != nil {
			return fmt.Errorf("registering font %s: %w", fontDef.Name, err)
		}

		// Add to current PDF instance
		info, _ := fontMgr.GetFontInfo(fontDef.Name)
		if info.FilePath != "" {
			if err := pdf.AddTTFFont(info.Name, info.FilePath); err != nil {
				return fmt.Errorf("adding font to PDF %s: %w", fontDef.Name, err)
			}
		}
	}

	return nil
}

// getPageSize returns the page size configuration
func (g *PDFGenerator) getPageSize(template *parser.DocumentTemplate) *gopdf.Rect {
	rect := &gopdf.Rect{}

	// Check for custom dimensions first
	if template.PageWidth > 0 && template.PageHeight > 0 {
		rect.W = template.PageWidth
		rect.H = template.PageHeight

		// Apply orientation
		if template.Orientation == "landscape" {
			rect.W, rect.H = rect.H, rect.W
		}
		return rect
	}

	// Use standard page sizes
	switch template.PageSize {
	case "A3":
		rect.W = gopdf.PageSizeA3.W
		rect.H = gopdf.PageSizeA3.H
	case "A5":
		rect.W = gopdf.PageSizeA5.W
		rect.H = gopdf.PageSizeA5.H
	case "Letter":
		rect.W = gopdf.PageSizeLetter.W
		rect.H = gopdf.PageSizeLetter.H
	case "Legal":
		rect.W = gopdf.PageSizeLegal.W
		rect.H = gopdf.PageSizeLegal.H
	default: // A4
		rect.W = gopdf.PageSizeA4.W
		rect.H = gopdf.PageSizeA4.H
	}

	// Apply orientation
	if template.Orientation == "landscape" {
		rect.W, rect.H = rect.H, rect.W
	}

	return rect
}

// getPageDimensions returns page width and height
func (g *PDFGenerator) getPageDimensions(template *parser.DocumentTemplate) (float64, float64) {
	rect := g.getPageSize(template)
	return rect.W, rect.H
}

// RegisterFont registers a font for use
func (g *PDFGenerator) RegisterFont(name, filePath string) error {
	return g.fontMgr.RegisterFont(name, filePath)
}

// RegisterFontFromBytes registers a font from byte data
func (g *PDFGenerator) RegisterFontFromBytes(name string, data []byte) error {
	return g.fontMgr.RegisterFontFromBytes(name, data)
}

// RegisterFontFromURL downloads and registers a font
func (g *PDFGenerator) RegisterFontFromURL(name, url string) error {
	return g.fontMgr.RegisterFontFromURL(name, url)
}

// LoadUnicodeFont loads a Unicode-capable font
func (g *PDFGenerator) LoadUnicodeFont(name, filePath string) error {
	return g.fontMgr.LoadUnicodeFont(name, filePath)
}

// LoadRTLFont loads an RTL-capable font
func (g *PDFGenerator) LoadRTLFont(name, filePath string) error {
	return g.fontMgr.LoadRTLFont(name, filePath)
}

// GetFontManager returns the font manager
func (g *PDFGenerator) GetFontManager() *fonts.Manager {
	return g.fontMgr
}

// Close cleans up resources
func (g *PDFGenerator) Close() error {
	// Cleanup temp files if needed
	return nil
}
