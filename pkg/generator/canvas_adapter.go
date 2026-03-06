package generator

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/amsaid/gopdf-generator/pkg/elements"
	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/signintech/gopdf"
)

// generateWithCanvas handles the "Canvas" mode (Absolute Positioning)
func (g *PDFGenerator) generateWithCanvas(template *parser.DocumentTemplate) (*bytes.Buffer, error) {
	pdf := &gopdf.GoPdf{}
	sessionFontMgr := g.fontMgr.Clone()

	// Initialize PDF
	if err := g.initializeCanvasPDF(pdf, template); err != nil {
		return nil, fmt.Errorf("initializing PDF: %w", err)
	}

	// Register base fonts
	if err := sessionFontMgr.AddFontsToPDF(pdf); err != nil {
		return nil, fmt.Errorf("adding base fonts: %w", err)
	}

	// Register template custom fonts
	if err := g.registerFonts(pdf, sessionFontMgr, template.Fonts); err != nil {
		return nil, fmt.Errorf("registering fonts: %w", err)
	}

	// Sort by ZIndex
	sort.SliceStable(template.Elements, func(i, j int) bool {
		return template.Elements[i].ZIndex < template.Elements[j].ZIndex
	})

	// Create Handler
	pageWidth, pageHeight := g.getPageDimensions(template)
	handler := elements.NewHandler(pdf, sessionFontMgr, pageWidth, pageHeight, template)

	// Draw Decorations
	if err := handler.DrawPageDecorations(); err != nil {
		return nil, fmt.Errorf("drawing decorations: %w", err)
	}

	// Process Elements
	for i, elem := range template.Elements {
		if err := handler.HandleElement(elem); err != nil {
			return nil, fmt.Errorf("processing element %d: %w", i, err)
		}
	}

	// Write
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("writing PDF: %w", err)
	}

	return &buf, nil
}

func (g *PDFGenerator) initializeCanvasPDF(pdf *gopdf.GoPdf, template *parser.DocumentTemplate) error {
	pageSize := g.getPageSize(template)
	pdf.Start(gopdf.Config{
		PageSize: *pageSize,
	})
	pdf.AddPage()
	pdf.SetInfo(gopdf.PdfInfo{
		Title:   template.Title,
		Author:  template.Author,
		Subject: template.Subject,
		Creator: template.Creator,
	})
	return nil
}

func (g *PDFGenerator) getPageDimensions(template *parser.DocumentTemplate) (float64, float64) {
	rect := g.getPageSize(template)
	return rect.W, rect.H
}
