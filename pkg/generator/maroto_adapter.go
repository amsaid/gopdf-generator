package generator

import (
	"bytes"
	"fmt"

	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/amsaid/gopdf-generator/pkg/rtl"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
)

// generateWithMaroto handles the "Flow" mode (Dynamic Layout)
func (g *PDFGenerator) generateWithMaroto(tmpl *parser.DocumentTemplate) (*bytes.Buffer, error) {
	// 1. Config
	cfgBuilder := config.NewBuilder().
		WithTopMargin(tmpl.Margin.Top).
		WithBottomMargin(tmpl.Margin.Bottom).
		WithLeftMargin(tmpl.Margin.Left).
		WithRightMargin(tmpl.Margin.Right)

	if tmpl.Orientation == "landscape" {
		cfgBuilder.WithOrientation(orientation.Horizontal)
	}
	if tmpl.PageSize == "Letter" {
		cfgBuilder.WithPageSize(pagesize.Letter)
	} else if tmpl.PageSize == "A3" {
		cfgBuilder.WithPageSize(pagesize.A3)
	} else if tmpl.PageSize == "Legal" {
		cfgBuilder.WithPageSize(pagesize.Legal)
	} else {
		cfgBuilder.WithPageSize(pagesize.A4)
	}

	// 2. Register Fonts (Hybrid: Find paths via our FontManager)
	registered := make(map[string]bool)
	fontRepo := repository.New()
	// Register system fonts from manager
	for _, fontName := range g.fontMgr.ListFonts() {
		info, exists := g.fontMgr.GetFontInfo(fontName)
		if exists && info.FilePath != "" && !registered[fontName] {
			fontRepo.AddUTF8Font(fontName, fontstyle.Normal, info.FilePath)
			fontRepo.AddUTF8Font(fontName, fontstyle.Bold, info.FilePath)
			registered[fontName] = true
		}
	}

	// Register template fonts
	for _, fDef := range tmpl.Fonts {
		if fDef.FilePath != "" && !registered[fDef.Name] {
			fontRepo.AddUTF8Font(fDef.Name, fontstyle.Normal, fDef.FilePath)
			fontRepo.AddUTF8Font(fDef.Name, fontstyle.Bold, fDef.FilePath)
			registered[fDef.Name] = true
		}
	}

	m := maroto.New(cfgBuilder.Build())

	// 3. Document Title (Optional Header)
	if tmpl.Title != "" {
		m.AddRow(20,
			col.New(12).Add(text.New(tmpl.Title, props.Text{
				Style: fontstyle.Bold,
				Size:  16,
				Align: align.Center,
			})),
		)
	}

	// 4. Build Flow Content
	for _, r := range tmpl.FlowContent {
		cols := g.buildMarotoColumns(r.Columns, tmpl.DefaultFont)

		rowHeight := r.Height
		if rowHeight == 0 {
			rowHeight = 10 // Default fallback
		}

		m.AddRow(rowHeight, cols...)
	}

	// 5. Generate
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("maroto generation failed: %w", err)
	}

	return bytes.NewBuffer(doc.GetBytes()), nil
}

func (g *PDFGenerator) buildMarotoColumns(cols []parser.FlowColumn, defaultFont *parser.FontConfig) []core.Col {
	var mCols []core.Col

	for _, c := range cols {
		size := c.Size
		if size < 1 {
			size = 1
		}
		if size > 12 {
			size = 12
		}

		newCol := col.New(size)

		// Text Component with RTL Injection
		if c.Text != "" {
			displayText := c.Text
			isRTL := rtl.ContainsRTL(displayText)

			// --- HYBRID RTL MAGIC ---
			if isRTL {
				displayText = rtl.ProcessRTLText(displayText)
			}

			// Style Props
			txtProps := props.Text{
				Size:   10,
				Align:  align.Left,
				Family: "Helvetica",
			}

			if defaultFont != nil {
				txtProps.Size = defaultFont.Size
				txtProps.Family = defaultFont.Family
			}

			if c.Style != nil {
				if c.Style.Font != nil {
					if c.Style.Font.Size > 0 {
						txtProps.Size = c.Style.Font.Size
					}
					if c.Style.Font.Family != "" {
						txtProps.Family = c.Style.Font.Family
					}
				}
				switch c.Style.Alignment {
				case "center":
					txtProps.Align = align.Center
				case "right":
					txtProps.Align = align.Right
				case "justify":
					txtProps.Align = align.Justify
				}
			}

			// Force Right alignment for RTL if not overridden
			if isRTL && (c.Style == nil || c.Style.Alignment == "") {
				txtProps.Align = align.Right
			}

			newCol.Add(text.New(displayText, txtProps))
		}

		// Image Component
		if c.Image != nil && c.Image.Path != "" {
			newCol.Add(image.NewFromFile(c.Image.Path))
		}

		mCols = append(mCols, newCol)
	}
	return mCols
}
