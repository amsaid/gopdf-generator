package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
		FontDir:    "../../fonts",
		TempDir:    os.TempDir(),
		EmbedFonts: true,
	}
}

// New creates a new PDF generator
func New(config *Config) (*PDFGenerator, error) {
	if config == nil {
		config = DefaultConfig()
	}

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

	// Hybrid Router: Switch based on Mode
	if template.Mode == "flow" {
		return g.generateWithMaroto(template)
	}

	// Default: Canvas Mode
	return g.generateWithCanvas(template)
}

// GenerateToFile generates a PDF and saves to file
func (g *PDFGenerator) GenerateToFile(template *parser.DocumentTemplate, outputPath string) error {
	buf, err := g.Generate(template)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
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

// --- Helper Functions ---

func (g *PDFGenerator) RegisterFont(name, filePath string) error {
	return g.fontMgr.RegisterFont(name, filePath)
}

func (g *PDFGenerator) RegisterFontFromBytes(name string, data []byte) error {
	return g.fontMgr.RegisterFontFromBytes(name, data)
}

func (g *PDFGenerator) LoadRTLFont(name, filePath string) error {
	return g.fontMgr.LoadRTLFont(name, filePath)
}

func (g *PDFGenerator) GetFontManager() *fonts.Manager {
	return g.fontMgr
}

func (g *PDFGenerator) Close() error {
	return nil
}

// shared helper for Canvas mode to load fonts
func (g *PDFGenerator) registerFonts(pdf *gopdf.GoPdf, fontMgr *fonts.Manager, fontDefs []parser.FontDef) error {
	for _, fontDef := range fontDefs {
		if fontDef.Name == "" {
			continue
		}
		if fontMgr.IsRegistered(fontDef.Name) {
			continue
		}

		var err error
		if len(fontDef.Data) > 0 {
			err = fontMgr.RegisterFontFromBytes(fontDef.Name, fontDef.Data)
		} else if fontDef.FilePath != "" {
			err = fontMgr.RegisterFont(fontDef.Name, fontDef.FilePath)
		}

		if err != nil {
			return fmt.Errorf("registering font %s: %w", fontDef.Name, err)
		}

		info, _ := fontMgr.GetFontInfo(fontDef.Name)
		if info.FilePath != "" {
			if err := pdf.AddTTFFont(info.Name, info.FilePath); err != nil {
				return fmt.Errorf("adding font to PDF %s: %w", fontDef.Name, err)
			}
		}
	}
	return nil
}

func (g *PDFGenerator) getPageSize(template *parser.DocumentTemplate) *gopdf.Rect {
	rect := &gopdf.Rect{}
	if template.PageWidth > 0 && template.PageHeight > 0 {
		rect.W = template.PageWidth
		rect.H = template.PageHeight
		if template.Orientation == "landscape" {
			rect.W, rect.H = rect.H, rect.W
		}
		return rect
	}
	// Defaults
	switch template.PageSize {
	case "A3":
		rect.W, rect.H = gopdf.PageSizeA3.W, gopdf.PageSizeA3.H
	case "Letter":
		rect.W, rect.H = gopdf.PageSizeLetter.W, gopdf.PageSizeLetter.H
	default: // A4
		rect.W, rect.H = gopdf.PageSizeA4.W, gopdf.PageSizeA4.H
	}
	if template.Orientation == "landscape" {
		rect.W, rect.H = rect.H, rect.W
	}
	return rect
}
