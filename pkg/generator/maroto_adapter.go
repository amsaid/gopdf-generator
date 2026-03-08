package generator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/amsaid/gopdf-generator/pkg/rtl"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/signature"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
)

// generateWithMaroto handles the "Flow" mode (Dynamic Layout) with full Maroto 2 support
func (g *PDFGenerator) generateWithMaroto(tmpl *parser.DocumentTemplate) (*bytes.Buffer, error) {
	// 1. Build Configuration with full Maroto 2 options
	cfgBuilder := g.buildMarotoConfig(tmpl)

	// 2. Register Fonts (Hybrid: Find paths via our FontManager)
	fontRepo := repository.New()
	registered := g.registerMarotoFonts(fontRepo, tmpl)

	// Add custom fonts from repository to builder
	customFonts, _ := fontRepo.Load()
	cfgBuilder.WithCustomFonts(customFonts)

	// 3. Create Maroto instance
	m := maroto.New(cfgBuilder.Build())
	m = g.applyMarotoMetadata(m, tmpl)

	// 4. Build Header (if specified in template)
	if len(tmpl.Header) > 0 {
		m = g.buildMarotoHeader(m, tmpl)
	}

	// 5. Build Footer (if specified in template)
	if len(tmpl.Footer) > 0 {
		m = g.buildMarotoFooter(m, tmpl)
	}

	// 6. Document Title (Optional)
	if tmpl.Title != "" {
		m.AddRows(g.buildTitleRow(tmpl))
	}

	// 7. Build Flow Content
	for _, r := range tmpl.FlowContent {
		cols := g.buildMarotoColumns(r.Columns, tmpl.DefaultFont, registered)

		rowHeight := r.Height
		if rowHeight == 0 {
			rowHeight = 10 // Default fallback
		}

		m.AddRow(rowHeight, cols...)
	}

	// 8. Add any additional table content
	for _, elem := range tmpl.Elements {
		if elem.Type == "table" {
			m.AddRows(g.buildMarotoTable(elem, tmpl.DefaultFont)...)
		}
	}

	// 9. Generate PDF
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("maroto generation failed: %w", err)
	}

	return bytes.NewBuffer(doc.GetBytes()), nil
}

// buildMarotoConfig creates a comprehensive Maroto configuration
func (g *PDFGenerator) buildMarotoConfig(tmpl *parser.DocumentTemplate) config.Builder {
	cfgBuilder := config.NewBuilder()

	// Page Margins
	cfgBuilder.WithTopMargin(tmpl.Margin.Top).
		WithBottomMargin(tmpl.Margin.Bottom).
		WithLeftMargin(tmpl.Margin.Left).
		WithRightMargin(tmpl.Margin.Right)

	// Page Size
	switch tmpl.PageSize {
	case "Letter":
		cfgBuilder.WithPageSize(pagesize.Letter)
	case "A3":
		cfgBuilder.WithPageSize(pagesize.A3)
	case "Legal":
		cfgBuilder.WithPageSize(pagesize.Legal)
	default:
		cfgBuilder.WithPageSize(pagesize.A4)
	}

	// Orientation
	if tmpl.Orientation == "landscape" {
		cfgBuilder.WithOrientation(orientation.Horizontal)
	}

	// Note: Maroto v2 doesn't have a global background color in Config.
	// Background colors can be set per Row/Col.

	// Protection (if needed)
	// cfgBuilder.WithProtection(protection.None, "user", "owner")

	// Compression
	cfgBuilder.WithCompression(true)

	// Page Numbering
	// cfgBuilder.WithPageNumber()

	// Debug Mode (grid)
	if tmpl.Grid != nil && tmpl.Grid.Draw {
		cfgBuilder.WithDebug(true)
	}

	return cfgBuilder
}

// applyMarotoMetadata applies document metadata
func (g *PDFGenerator) applyMarotoMetadata(m core.Maroto, tmpl *parser.DocumentTemplate) core.Maroto {
	if tmpl.Title != "" || tmpl.Author != "" || tmpl.Subject != "" || tmpl.Creator != "" {
		// Note: Maroto 2 handles metadata through config
		// Additional metadata can be set via PDF properties
	}
	return m
}

// registerMarotoFonts registers fonts for use in Maroto
func (g *PDFGenerator) registerMarotoFonts(fontRepo repository.Repository, tmpl *parser.DocumentTemplate) map[string]bool {
	registered := make(map[string]bool)

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

	return registered
}

// buildMarotoHeader creates header rows
func (g *PDFGenerator) buildMarotoHeader(m core.Maroto, tmpl *parser.DocumentTemplate) core.Maroto {
	var headerRows []core.Row

	for _, elem := range tmpl.Header {
		if elem.Type == "text" {
			rowHeight := 10.0
			if elem.Height > 0 {
				rowHeight = elem.Height
			}

			txt := elem.Text
			isRTL := rtl.ContainsRTL(txt)
			if isRTL {
				txt = rtl.ProcessRTLText(txt)
			}

			txtProps := g.buildTextProps(elem, tmpl.DefaultFont)
			if isRTL {
				txtProps.Align = align.Right
			}

			headerRows = append(headerRows, row.New(rowHeight).Add(
				text.NewCol(12, txt, txtProps),
			))
		}
	}

	if len(headerRows) > 0 {
		m.RegisterHeader(headerRows...)
	}

	return m
}

// buildMarotoFooter creates footer rows
func (g *PDFGenerator) buildMarotoFooter(m core.Maroto, tmpl *parser.DocumentTemplate) core.Maroto {
	var footerRows []core.Row

	for _, elem := range tmpl.Footer {
		if elem.Type == "text" {
			rowHeight := 10.0
			if elem.Height > 0 {
				rowHeight = elem.Height
			}

			txt := elem.Text
			isRTL := rtl.ContainsRTL(txt)
			if isRTL {
				txt = rtl.ProcessRTLText(txt)
			}

			txtProps := g.buildTextProps(elem, tmpl.DefaultFont)
			if isRTL {
				txtProps.Align = align.Right
			}

			footerRows = append(footerRows, row.New(rowHeight).Add(
				text.NewCol(12, txt, txtProps),
			))
		}
	}

	if len(footerRows) > 0 {
		m.RegisterFooter(footerRows...)
	}

	return m
}

// buildTitleRow creates a title row
func (g *PDFGenerator) buildTitleRow(tmpl *parser.DocumentTemplate) core.Row {
	return row.New(20).Add(
		text.NewCol(12, tmpl.Title, props.Text{
			Style: fontstyle.Bold,
			Size:  16,
			Align: align.Center,
		}),
	)
}

// buildMarotoColumns builds columns from flow column definitions
func (g *PDFGenerator) buildMarotoColumns(cols []parser.FlowColumn, defaultFont *parser.FontConfig, registeredFonts map[string]bool) []core.Col {
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

		// Apply column style (background, border, etc.)
		if c.Style != nil {
			newCol = g.applyColumnStyle(newCol, c.Style)
		}

		// Process nested components
		if len(c.Components) > 0 {
			for _, comp := range c.Components {
				g.addComponentToCol(newCol, comp, defaultFont, registeredFonts)
			}
		} else if c.Text != "" {
			// Text Component with RTL Injection
			g.addTextToCol(newCol, c.Text, c.Style, defaultFont)
		}

		// Image Component
		if c.Image != nil {
			g.addImageToCol(newCol, c.Image, c.Style)
		}

		mCols = append(mCols, newCol)
	}
	return mCols
}

// applyColumnStyle applies styling to a column
func (g *PDFGenerator) applyColumnStyle(c core.Col, style *parser.FlowStyle) core.Col {
	if style == nil {
		return c
	}

	cellStyle := &props.Cell{}

	// Background color
	if style.BgColor != nil {
		cellStyle.BackgroundColor = &props.Color{
			Red:   int(style.BgColor.R),
			Green: int(style.BgColor.G),
			Blue:  int(style.BgColor.B),
		}
	}

	// Border
	if style.BorderColor != nil || style.BorderWidth > 0 {
		cellStyle.BorderType = border.Full
		cellStyle.BorderThickness = style.BorderWidth
		if style.BorderColor != nil {
			cellStyle.BorderColor = &props.Color{
				Red:   int(style.BorderColor.R),
				Green: int(style.BorderColor.G),
				Blue:  int(style.BorderColor.B),
			}
		}

		if style.BorderType == "dashed" {
			cellStyle.LineStyle = linestyle.Dashed
		} else {
			cellStyle.LineStyle = linestyle.Solid
		}
	}

	// Note: Maroto v2.3.3 props.Cell does not have Padding fields.
	// Padding is often handled by adding empty cols/rows or via component specific props if available.

	return c.WithStyle(cellStyle)
}

// addComponentToCol adds a component to a column based on type
func (g *PDFGenerator) addComponentToCol(c core.Col, comp parser.FlowComponent, defaultFont *parser.FontConfig, registeredFonts map[string]bool) {
	switch comp.Type {
	case "text":
		g.addTextToCol(c, comp.Text, comp.Style, defaultFont)

	case "image":
		if comp.Image != nil {
			g.addImageToCol(c, comp.Image, comp.Style)
		}

	case "signature":
		g.addSignatureToCol(c, comp.Text, comp.Style)

	case "barcode":
		g.addBarcodeToCol(c, comp.Text, comp.Style)

	case "qrcode":
		g.addQRCodeToCol(c, comp.Text, comp.Style)

	case "line":
		g.addLineToCol(c, comp.Style)
	}
}

// addTextToCol adds text to a column with RTL support
func (g *PDFGenerator) addTextToCol(c core.Col, txt string, style *parser.FlowStyle, defaultFont *parser.FontConfig) {
	displayText := txt
	isRTL := rtl.ContainsRTL(displayText)

	// Process RTL text
	if isRTL {
		displayText = rtl.ProcessRTLText(displayText)
	}

	txtProps := props.Text{
		Size:   10,
		Align:  align.Left,
		Family: "Helvetica",
	}

	if defaultFont != nil {
		txtProps.Size = defaultFont.Size
		txtProps.Family = defaultFont.Family
	}

	if style != nil {
		if style.Font != nil {
			if style.Font.Size > 0 {
				txtProps.Size = style.Font.Size
			}
			if style.Font.Family != "" {
				txtProps.Family = style.Font.Family
			}
			if style.Font.Style != "" {
				txtProps.Style = g.parseMarotoFontStyle(style.Font.Style)
			}
			if style.Font.Color != nil {
				txtProps.Color = &props.Color{
					Red:   int(style.Font.Color.R),
					Green: int(style.Font.Color.G),
					Blue:  int(style.Font.Color.B),
				}
			}
		}

		switch style.Alignment {
		case "center":
			txtProps.Align = align.Center
		case "right":
			txtProps.Align = align.Right
		case "justify":
			txtProps.Align = align.Justify
		}

		// Text color
		if style.TextColor != nil {
			txtProps.Color = &props.Color{
				Red:   int(style.TextColor.R),
				Green: int(style.TextColor.G),
				Blue:  int(style.TextColor.B),
			}
		}
	}

	// Force Right alignment for RTL if not overridden
	if isRTL && (style == nil || style.Alignment == "") {
		txtProps.Align = align.Right
	}

	c.Add(text.New(displayText, txtProps))
}

// addImageToCol adds an image to a column
func (g *PDFGenerator) addImageToCol(c core.Col, img *parser.FlowImage, style *parser.FlowStyle) {
	// Handle different image sources
	if img.Path != "" {
		imgProps := props.Rect{}
		if img.Width > 0 {
			imgProps.Percent = img.Width // Approximate mapping
		}

		c.Add(image.NewFromFile(img.Path, imgProps))

	} else if img.Base64 != "" {
		data, err := base64.StdEncoding.DecodeString(img.Base64)
		if err != nil {
			return
		}

		ext := extension.Png
		if img.Extension != "" {
			ext = g.parseExtension(img.Extension)
		}

		imgProps := props.Rect{}
		if img.Width > 0 {
			imgProps.Percent = img.Width
		}

		c.Add(image.NewFromBytes(data, ext, imgProps))

	} else if img.URL != "" {
		// Download image from URL
		data, ext, err := g.downloadImage(img.URL)
		if err != nil {
			return
		}

		imgProps := props.Rect{}
		if img.Width > 0 {
			imgProps.Percent = img.Width
		}

		c.Add(image.NewFromBytes(data, ext, imgProps))
	}
}

// addSignatureToCol adds a signature to a column
func (g *PDFGenerator) addSignatureToCol(c core.Col, label string, style *parser.FlowStyle) {
	sigProps := props.Signature{
		FontFamily: "Helvetica",
		FontStyle:  fontstyle.Normal,
		FontSize:   10,
	}

	if style != nil && style.Font != nil {
		if style.Font.Family != "" {
			sigProps.FontFamily = style.Font.Family
		}
		if style.Font.Size > 0 {
			sigProps.FontSize = style.Font.Size
		}
		if style.Font.Style != "" {
			sigProps.FontStyle = g.parseMarotoFontStyle(style.Font.Style)
		}
		if style.Font.Color != nil {
			sigProps.FontColor = &props.Color{
				Red:   int(style.Font.Color.R),
				Green: int(style.Font.Color.G),
				Blue:  int(style.Font.Color.B),
			}
		}
	}

	c.Add(signature.New(label, sigProps))
}

// addBarcodeToCol adds a barcode to a column
func (g *PDFGenerator) addBarcodeToCol(c core.Col, codeStr string, style *parser.FlowStyle) {
	barProps := props.Barcode{
		Percent: 100,
	}

	// Note: props.Barcode in Maroto v2.3.3 does not have a Color field
	c.Add(code.NewBar(codeStr, barProps))
}

// addQRCodeToCol adds a QR code to a column
func (g *PDFGenerator) addQRCodeToCol(c core.Col, codeStr string, style *parser.FlowStyle) {
	qrProps := props.Rect{
		Percent: 100,
	}

	// Note: props.Rect in Maroto v2.3.3 does not have a Color field
	c.Add(code.NewQr(codeStr, qrProps))
}

// addLineToCol adds a line to a column
func (g *PDFGenerator) addLineToCol(c core.Col, style *parser.FlowStyle) {
	lineProps := props.Line{
		Style:     linestyle.Solid,
		Thickness: 1.0,
		Color:     &props.Color{Red: 0, Green: 0, Blue: 0},
	}

	if style != nil {
		if style.BorderType == "dashed" {
			lineProps.Style = linestyle.Dashed
		}
		if style.BorderWidth > 0 {
			lineProps.Thickness = style.BorderWidth
		}
		if style.BorderColor != nil {
			lineProps.Color = &props.Color{
				Red:   int(style.BorderColor.R),
				Green: int(style.BorderColor.G),
				Blue:  int(style.BorderColor.B),
			}
		}
	}

	c.Add(line.New(lineProps))
}

// buildMarotoTable creates table rows from table element
func (g *PDFGenerator) buildMarotoTable(elem parser.Element, defaultFont *parser.FontConfig) []core.Row {
	var rows []core.Row

	// Calculate column sizes
	colSizes := g.calculateMarotoColumnSizes(elem)

	// Add header
	if elem.Header != nil {
		rows = append(rows, g.buildMarotoTableHeader(elem.Header, colSizes, elem))
	}

	// Add rows
	for _, rowData := range elem.Rows {
		rows = append(rows, g.buildMarotoTableRow(rowData, colSizes, elem, defaultFont))
	}

	// Add footer
	if elem.Footer != nil {
		rows = append(rows, g.buildMarotoTableFooter(elem.Footer, colSizes, elem))
	}

	return rows
}

// calculateMarotoColumnSizes calculates column sizes for Maroto (1-12 grid)
func (g *PDFGenerator) calculateMarotoColumnSizes(elem parser.Element) []int {
	sizes := make([]int, len(elem.Columns))
	total := 0

	for i, col := range elem.Columns {
		if col.Width > 0 {
			// Normalize to 1-12 scale
			sizes[i] = int(col.Width / 50) // Approximate conversion
			if sizes[i] < 1 {
				sizes[i] = 1
			}
			if sizes[i] > 12 {
				sizes[i] = 12
			}
		} else {
			sizes[i] = 1
		}
		total += sizes[i]
	}

	// Normalize to fit 12 columns total
	if total > 12 {
		scale := 12.0 / float64(total)
		for i := range sizes {
			sizes[i] = int(float64(sizes[i]) * scale)
			if sizes[i] < 1 {
				sizes[i] = 1
			}
		}
	}

	return sizes
}

// buildMarotoTableHeader creates table header row
func (g *PDFGenerator) buildMarotoTableHeader(header *parser.TableSection, colSizes []int, elem parser.Element) core.Row {
	r := row.New(10)

	cellStyle := props.Cell{
		BackgroundColor: &props.Color{Red: 52, Green: 73, Blue: 94},
	}

	if header.Background != nil {
		cellStyle.BackgroundColor = &props.Color{
			Red:   int(header.Background.R),
			Green: int(header.Background.G),
			Blue:  int(header.Background.B),
		}
	}

	for i, cell := range header.Cells {
		size := colSizes[i]
		if i >= len(colSizes) {
			size = 1
		}

		txtProps := props.Text{
			Style: fontstyle.Bold,
			Size:  10,
			Color: &props.Color{Red: 255, Green: 255, Blue: 255},
		}

		if cell.Font != nil {
			if cell.Font.Style != "" {
				txtProps.Style = g.parseMarotoFontStyle(cell.Font.Style)
			}
			if cell.Font.Size > 0 {
				txtProps.Size = cell.Font.Size
			}
			if cell.Font.Color != nil {
				txtProps.Color = &props.Color{
					Red:   int(cell.Font.Color.R),
					Green: int(cell.Font.Color.G),
					Blue:  int(cell.Font.Color.B),
				}
			}
		}

		c := col.New(size).WithStyle(&cellStyle)
		c.Add(text.New(cell.Text, txtProps))
		r.Add(c)
	}

	return r
}

// buildMarotoTableRow creates a table data row
func (g *PDFGenerator) buildMarotoTableRow(rowData parser.TableRow, colSizes []int, elem parser.Element, defaultFont *parser.FontConfig) core.Row {
	rowHeight := 8.0
	if rowData.Height > 0 {
		rowHeight = rowData.Height
	}
	if elem.MinRowHeight > 0 && rowHeight < elem.MinRowHeight {
		rowHeight = elem.MinRowHeight
	}

	r := row.New(rowHeight)

	cellStyle := props.Cell{}
	if rowData.Background != nil {
		cellStyle.BackgroundColor = &props.Color{
			Red:   int(rowData.Background.R),
			Green: int(rowData.Background.G),
			Blue:  int(rowData.Background.B),
		}
	}

	for i, cell := range rowData.Cells {
		size := colSizes[i]
		if i >= len(colSizes) {
			size = 1
		}

		txtProps := props.Text{
			Size:  10,
			Align: align.Left,
		}

		if defaultFont != nil {
			txtProps.Family = defaultFont.Family
			txtProps.Size = defaultFont.Size
		}

		if cell.Font != nil {
			if cell.Font.Family != "" {
				txtProps.Family = cell.Font.Family
			}
			if cell.Font.Size > 0 {
				txtProps.Size = cell.Font.Size
			}
			if cell.Font.Style != "" {
				txtProps.Style = g.parseMarotoFontStyle(cell.Font.Style)
			}
			if cell.Font.Color != nil {
				txtProps.Color = &props.Color{
					Red:   int(cell.Font.Color.R),
					Green: int(cell.Font.Color.G),
					Blue:  int(cell.Font.Color.B),
				}
			}
		}

		// Alignment
		switch cell.Align {
		case "C", "center":
			txtProps.Align = align.Center
		case "R", "right":
			txtProps.Align = align.Right
		case "L", "left":
			txtProps.Align = align.Left
		}

		// RTL support
		txt := cell.Text
		if cell.RTL || rtl.ContainsRTL(txt) {
			txt = rtl.ProcessRTLText(txt)
			if cell.Align == "" {
				txtProps.Align = align.Right
			}
		}

		// Cell-specific background
		cellBgStyle := cellStyle
		if cell.Background != nil {
			cellBgStyle = props.Cell{
				BackgroundColor: &props.Color{
					Red:   int(cell.Background.R),
					Green: int(cell.Background.G),
					Blue:  int(cell.Background.B),
				},
			}
		}

		c := col.New(size).WithStyle(&cellBgStyle)
		c.Add(text.New(txt, txtProps))
		r.Add(c)
	}

	return r
}

// buildMarotoTableFooter creates table footer row
func (g *PDFGenerator) buildMarotoTableFooter(footer *parser.TableSection, colSizes []int, elem parser.Element) core.Row {
	r := row.New(10)

	cellStyle := props.Cell{
		BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
	}

	if footer.Background != nil {
		cellStyle.BackgroundColor = &props.Color{
			Red:   int(footer.Background.R),
			Green: int(footer.Background.G),
			Blue:  int(footer.Background.B),
		}
	}

	for i, cell := range footer.Cells {
		size := colSizes[i]
		if i >= len(colSizes) {
			size = 1
		}

		txtProps := props.Text{
			Style: fontstyle.Bold,
			Size:  10,
		}

		if cell.Font != nil {
			if cell.Font.Style != "" {
				txtProps.Style = g.parseMarotoFontStyle(cell.Font.Style)
			}
			if cell.Font.Size > 0 {
				txtProps.Size = cell.Font.Size
			}
			if cell.Font.Color != nil {
				txtProps.Color = &props.Color{
					Red:   int(cell.Font.Color.R),
					Green: int(cell.Font.Color.G),
					Blue:  int(cell.Font.Color.B),
				}
			}
		}

		c := col.New(size).WithStyle(&cellStyle)
		c.Add(text.New(cell.Text, txtProps))
		r.Add(c)
	}

	return r
}

// Helper functions

func (g *PDFGenerator) buildTextProps(elem parser.Element, defaultFont *parser.FontConfig) props.Text {
	txtProps := props.Text{
		Size:   10,
		Align:  align.Left,
		Family: "Helvetica",
	}

	if defaultFont != nil {
		txtProps.Size = defaultFont.Size
		txtProps.Family = defaultFont.Family
	}

	if elem.Font != nil {
		if elem.Font.Size > 0 {
			txtProps.Size = elem.Font.Size
		}
		if elem.Font.Family != "" {
			txtProps.Family = elem.Font.Family
		}
		if elem.Font.Style != "" {
			txtProps.Style = g.parseMarotoFontStyle(elem.Font.Style)
		}
		if elem.Font.Color != nil {
			txtProps.Color = &props.Color{
				Red:   int(elem.Font.Color.R),
				Green: int(elem.Font.Color.G),
				Blue:  int(elem.Font.Color.B),
			}
		}
	}

	if elem.Alignment != nil {
		switch elem.Alignment.Horizontal {
		case "C", "center":
			txtProps.Align = align.Center
		case "R", "right":
			txtProps.Align = align.Right
		case "L", "left":
			txtProps.Align = align.Left
		case "J", "justify":
			txtProps.Align = align.Justify
		}
	}

	return txtProps
}

func (g *PDFGenerator) parseMarotoFontStyle(style string) fontstyle.Type {
	style = strings.ToUpper(style)
	switch style {
	case "B", "BOLD":
		return fontstyle.Bold
	case "I", "ITALIC":
		return fontstyle.Italic
	case "BI", "BOLDITALIC":
		return fontstyle.BoldItalic
	default:
		return fontstyle.Normal
	}
}

func (g *PDFGenerator) parseExtension(ext string) extension.Type {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg":
		return extension.Jpg
	case "png":
		return extension.Png
	default:
		return extension.Png
	}
}

func (g *PDFGenerator) getExtensionFromPath(path string) extension.Type {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return extension.Jpg
	case ".png":
		return extension.Png
	default:
		return extension.Png
	}
}

func (g *PDFGenerator) downloadImage(url string) ([]byte, extension.Type, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, extension.Png, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, extension.Png, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, extension.Png, err
	}

	// Detect extension from content type
	contentType := resp.Header.Get("Content-Type")
	ext := extension.Png
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		ext = extension.Jpg
	}

	return data, ext, nil
}
