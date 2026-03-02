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
	pdf     *gopdf.GoPdf
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

	pdf := &gopdf.GoPdf{}

	fontMgr := fonts.NewManager(pdf, config.FontDir)
	fontMgr.RegisterStandardFonts()

	return &PDFGenerator{
		pdf:     pdf,
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

	// Initialize PDF
	if err := g.initializePDF(template); err != nil {
		return nil, fmt.Errorf("initializing PDF: %w", err)
	}

	// Register custom fonts
	if err := g.registerFonts(template.Fonts); err != nil {
		return nil, fmt.Errorf("registering fonts: %w", err)
	}

	// Create element handler
	pageWidth, pageHeight := g.getPageDimensions(template)
	handler := elements.NewHandler(g.pdf, g.fontMgr, pageWidth, pageHeight, template.Margin, template.DefaultFont)

	// Process elements
	for i, elem := range template.Elements {
		if err := handler.HandleElement(elem); err != nil {
			return nil, fmt.Errorf("processing element %d (%s): %w", i, elem.Type, err)
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := g.pdf.Write(&buf); err != nil {
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
func (g *PDFGenerator) initializePDF(template *parser.DocumentTemplate) error {
	// Set page size
	pageSize := g.getPageSize(template)

	// Start PDF with page size
	g.pdf.Start(gopdf.Config{
		PageSize: pageSize,
	})

	// Add first page
	g.pdf.AddPage()

	// Set metadata
	if template.Title != "" {
		g.pdf.SetTitle(template.Title)
	}
	if template.Author != "" {
		g.pdf.SetAuthor(template.Author)
	}
	if template.Subject != "" {
		g.pdf.SetSubject(template.Subject)
	}
	if template.Creator != "" {
		g.pdf.SetCreator(template.Creator)
	}

	return nil
}

// registerFonts registers custom fonts from template
func (g *PDFGenerator) registerFonts(fontDefs []parser.FontDef) error {
	for _, fontDef := range fontDefs {
		if fontDef.Name == "" {
			continue
		}

		// If font data provided, use it
		if len(fontDef.Data) > 0 {
			if err := g.fontMgr.RegisterFontFromBytes(fontDef.Name, fontDef.Data); err != nil {
				return fmt.Errorf("registering font %s from bytes: %w", fontDef.Name, err)
			}
			continue
		}

		// Otherwise use file path
		if fontDef.FilePath != "" {
			if err := g.fontMgr.RegisterFont(fontDef.Name, fontDef.FilePath); err != nil {
				return fmt.Errorf("registering font %s from file: %w", fontDef.Name, err)
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
		rect.W = gopdf.A3Size.W
		rect.H = gopdf.A3Size.H
	case "A5":
		rect.W = gopdf.A5Size.W
		rect.H = gopdf.A5Size.H
	case "Letter":
		rect.W = gopdf.LetterSize.W
		rect.H = gopdf.LetterSize.H
	case "Legal":
		rect.W = gopdf.LegalSize.W
		rect.H = gopdf.LegalSize.H
	default: // A4
		rect.W = gopdf.A4Size.W
		rect.H = gopdf.A4Size.H
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
