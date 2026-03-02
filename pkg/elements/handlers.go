package elements

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/amsaid/gopdf-generator/pkg/fonts"
	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/amsaid/gopdf-generator/pkg/rtl"
	"github.com/signintech/gopdf"
)

// Handler handles PDF element rendering
type Handler struct {
	pdf         *gopdf.GoPdf
	fontMgr     *fonts.Manager
	cursorY     float64
	pageWidth   float64
	pageHeight  float64
	margin      *parser.Margin
	defaultFont *parser.FontConfig
}

// NewHandler creates a new element handler
func NewHandler(pdf *gopdf.GoPdf, fontMgr *fonts.Manager, pageWidth, pageHeight float64, margin *parser.Margin, defaultFont *parser.FontConfig) *Handler {
	return &Handler{
		pdf:         pdf,
		fontMgr:     fontMgr,
		cursorY:     margin.Top,
		pageWidth:   pageWidth,
		pageHeight:  pageHeight,
		margin:      margin,
		defaultFont: defaultFont,
	}
}

// SetCursorY sets the current Y position
func (h *Handler) SetCursorY(y float64) {
	h.cursorY = y
}

// GetCursorY returns the current Y position
func (h *Handler) GetCursorY() float64 {
	return h.cursorY
}

// CheckPageBreak checks if we need a page break
func (h *Handler) CheckPageBreak(requiredHeight float64) error {
	if h.cursorY+requiredHeight > h.pageHeight-h.margin.Bottom {
		h.pdf.AddPage()
		h.cursorY = h.margin.Top
	}
	return nil
}

// HandleElement processes a single element
func (h *Handler) HandleElement(elem parser.Element) error {
	switch elem.Type {
	case "text":
		return h.handleText(elem)
	case "cell":
		return h.handleCell(elem)
	case "image":
		return h.handleImage(elem)
	case "table":
		return h.handleTable(elem)
	case "line":
		return h.handleLine(elem)
	case "rect", "rectangle":
		return h.handleRect(elem)
	case "ellipse":
		return h.handleEllipse(elem)
	case "newline", "br":
		return h.handleNewline(elem)
	case "pagebreak":
		return h.handlePageBreak()
	case "list":
		return h.handleList(elem)
	case "link":
		return h.handleLink(elem)
	default:
		return fmt.Errorf("unknown element type: %s", elem.Type)
	}
}

// handleText renders text element
func (h *Handler) handleText(elem parser.Element) error {
	font := h.getFontConfig(elem.Font)

	// Set font
	style := font.Style
	if style == "" {
		style = ""
	}

	if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

	// Set color
	if font.Color != nil {
		h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
	} else {
		h.pdf.SetTextColor(0, 0, 0)
	}

	// Process RTL text (Shape the text FIRST to guarantee exact logical wrap lengths)
	text := elem.Text
	isRTL := elem.RTL || rtl.IsRTLText(text)
	if isRTL || rtl.ContainsRTL(text) {
		text = rtl.ShapeArabic(text)
	}

	// Get position
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	// Calculate text dimensions
	width := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Size != nil && elem.Size.Width > 0 {
		width = elem.Size.Width
	}

	// Handle alignment
	align := "L"
	if isRTL {
		align = "R" // default for RTL text
	}
	if elem.Alignment != nil && elem.Alignment.Horizontal != "" {
		align = elem.Alignment.Horizontal
	}

	// Calculate line height
	lineHeight := font.Size * 1.2
	if elem.LineHeight > 0 {
		lineHeight = elem.LineHeight
	}

	// Split text into lines (Words stay logically ordered here)
	lines := h.wrapText(text, width, font)

	// Check page break
	totalHeight := float64(len(lines)) * lineHeight
	if err := h.CheckPageBreak(totalHeight); err != nil {
		return err
	}

	// Render each line
	for _, line := range lines {
		lineX := x

		// Setup printable string. Reverse the logically bound line into final visual string
		renderLine := line
		if isRTL || rtl.ContainsRTL(line) {
			renderLine = rtl.ReorderString(line, isRTL)
		}

		// Apply alignment
		switch align {
		case "C", "center":
			lineWidth, _ := h.pdf.MeasureTextWidth(renderLine)
			lineX = x + (width-lineWidth)/2
		case "R", "right":
			lineWidth, _ := h.pdf.MeasureTextWidth(renderLine)
			lineX = x + width - lineWidth
		}

		h.pdf.SetXY(lineX, y)
		h.pdf.Cell(nil, renderLine)

		y += lineHeight
	}

	// Update cursor
	if elem.Position == nil {
		h.cursorY = y
	}

	return nil
}

// handleList renders a list (bulleted or numbered) element
func (h *Handler) handleList(elem parser.Element) error {
	font := h.getFontConfig(elem.Font)

	if err := h.fontMgr.SetFont(h.pdf, font.Family, font.Style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

	if font.Color != nil {
		h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
	} else {
		h.pdf.SetTextColor(0, 0, 0)
	}

	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	width := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Size != nil && elem.Size.Width > 0 {
		width = elem.Size.Width
	}

	lineHeight := font.Size * 1.2
	if elem.LineHeight > 0 {
		lineHeight = elem.LineHeight
	}
	bulletIndent := font.Size * 1.5 // Space reserved for the bullet point

	for i, item := range elem.ListItems {
		// Determine bullet marker
		bullet := "•"
		if elem.ListType == "ol" {
			bullet = fmt.Sprintf("%d.", i+1)
		}

		isRTL := elem.RTL || rtl.IsRTLText(item)
		renderItem := item
		if isRTL || rtl.ContainsRTL(item) {
			renderItem = rtl.ShapeArabic(item)
		}

		lines := h.wrapText(renderItem, width-bulletIndent, font)

		if err := h.CheckPageBreak(lineHeight * float64(len(lines))); err != nil {
			return err
		}

		// Draw bullet
		bulletX := x
		if isRTL {
			// In RTL, bullet sits on the right
			bulletX = x + width - bulletIndent + 5
		}
		h.pdf.SetXY(bulletX, y)
		h.pdf.Cell(nil, bullet)

		// Draw lines
		for _, line := range lines {
			renderLine := line
			if isRTL || rtl.ContainsRTL(line) {
				renderLine = rtl.ReorderString(line, isRTL)
			}

			lineX := x + bulletIndent
			if isRTL {
				lineWidth, _ := h.pdf.MeasureTextWidth(renderLine)
				lineX = x + width - bulletIndent - lineWidth
			}

			h.pdf.SetXY(lineX, y)
			h.pdf.Cell(nil, renderLine)
			y += lineHeight
		}
		y += lineHeight * 0.3 // minor padding between items
	}

	if elem.Position == nil {
		h.cursorY = y
	}

	return nil
}

// handleLink renders a clickable hyperlink
func (h *Handler) handleLink(elem parser.Element) error {
	font := h.getFontConfig(elem.Font)

	// Ensure links are underlined by default
	style := font.Style
	if !strings.Contains(strings.ToUpper(style), "U") {
		style += "U"
	}

	if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

	// Default link color: Standard blue if not specified
	if font.Color != nil {
		h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
	} else {
		h.pdf.SetTextColor(17, 85, 204)
	}

	text := elem.Text
	isRTL := elem.RTL || rtl.IsRTLText(text)
	if isRTL || rtl.ContainsRTL(text) {
		text = rtl.ShapeArabic(text)
	}

	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	renderText := text
	if isRTL || rtl.ContainsRTL(text) {
		renderText = rtl.ReorderString(text, isRTL)
	}

	w, _ := h.pdf.MeasureTextWidth(renderText)

	if err := h.CheckPageBreak(font.Size * 1.5); err != nil {
		return err
	}

	h.pdf.SetXY(x, y)
	h.pdf.Cell(nil, renderText)
	h.pdf.AddExternalLink(elem.URL, x, y, w, font.Size)

	if elem.Position == nil {
		h.cursorY = y + font.Size*1.5
	}

	return nil
}

// handleCell renders a cell element
func (h *Handler) handleCell(elem parser.Element) error {
	font := h.getFontConfig(elem.Font)

	// Set font
	style := font.Style
	if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

	// Set colors
	if font.Color != nil {
		h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
	} else {
		h.pdf.SetTextColor(0, 0, 0)
	}

	// Get position and size
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	width := h.pageWidth - h.margin.Left - h.margin.Right
	height := font.Size * 1.5
	if elem.Size != nil {
		if elem.Size.Width > 0 {
			width = elem.Size.Width
		}
		if elem.Size.Height > 0 {
			height = elem.Size.Height
		}
	}

	// Check page break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}

	// Draw background
	if elem.BackgroundColor != nil {
		h.pdf.SetFillColor(elem.BackgroundColor.R, elem.BackgroundColor.G, elem.BackgroundColor.B)
		h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, "F")
	}

	// Draw border
	if elem.Border != nil {
		h.drawBorder(x, y, width, height, elem.Border, elem.BorderColor)
	}

	// Process text
	text := elem.Text
	if elem.RTL || rtl.IsRTLText(text) {
		text = rtl.ProcessRTLText(text)
	}

	// Apply alignment
	alignStr := "LT"
	if elem.RTL || rtl.IsRTLText(elem.Text) {
		alignStr = "RT"
	}
	if elem.Alignment != nil {
		switch elem.Alignment.Horizontal {
		case "C", "center":
			alignStr = "CT"
		case "R", "right":
			alignStr = "RT"
		case "L", "left":
			alignStr = "LT"
		}
		switch elem.Alignment.Vertical {
		case "M", "middle":
			alignStr = alignStr[:1] + "M"
		case "B", "bottom":
			alignStr = alignStr[:1] + "B"
		}
	}

	// Convert alignment string to int constant
	align := h.parseAlign(alignStr)

	// Draw cell with text
	h.pdf.SetXY(x, y)
	h.pdf.CellWithOption(&gopdf.Rect{
		W: width,
		H: height,
	}, text, gopdf.CellOption{
		Align:  align,
		Border: 0,
		Float:  gopdf.Left,
	})

	// Update cursor
	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// handleImage renders an image element
func (h *Handler) handleImage(elem parser.Element) error {
	var imagePath string
	var cleanup bool

	// Warn if the user passes an SVG, as gopdf only natively supports JPG/PNG.
	if strings.HasSuffix(strings.ToLower(elem.ImagePath), ".svg") || strings.HasPrefix(elem.ImageURL, "data:image/svg") {
		return fmt.Errorf("SVG images are not natively supported by the PDF engine. Please convert to PNG or JPEG")
	}

	// Get image data
	if elem.ImagePath != "" {
		imagePath = elem.ImagePath
	} else if len(elem.ImageData) > 0 {
		// Save to temp file
		tempFile, err := os.CreateTemp("", "gopdf-img-*.png")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())

		if _, err := tempFile.Write(elem.ImageData); err != nil {
			tempFile.Close()
			return fmt.Errorf("writing image data: %w", err)
		}
		tempFile.Close()
		imagePath = tempFile.Name()
		cleanup = true
	} else if strings.HasPrefix(elem.ImageURL, "data:image/") {
		// Handle embedded Base64 data URIs
		parts := strings.SplitN(elem.ImageURL, ",", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid base64 image URL")
		}
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("decoding base64 image: %w", err)
		}
		tempFile, err := os.CreateTemp("", "gopdf-img-*")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())

		if _, err := tempFile.Write(data); err != nil {
			tempFile.Close()
			return fmt.Errorf("writing base64 image data: %w", err)
		}
		tempFile.Close()
		imagePath = tempFile.Name()
		cleanup = true
	} else if elem.ImageURL != "" {
		// Download image
		resp, err := http.Get(elem.ImageURL)
		if err != nil {
			return fmt.Errorf("downloading image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("downloading image: status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading image data: %w", err)
		}

		// Save to temp file
		tempFile, err := os.CreateTemp("", "gopdf-img-*")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())

		if _, err := tempFile.Write(data); err != nil {
			tempFile.Close()
			return fmt.Errorf("writing image data: %w", err)
		}
		tempFile.Close()
		imagePath = tempFile.Name()
		cleanup = true
	}

	// Always cleanup downloaded/base64 extracted temp files once the handler completes
	if cleanup {
		defer os.Remove(imagePath)
	}

	// Get position
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	// Get dimensions
	width := 0.0
	height := 0.0
	if elem.Size != nil {
		width = elem.Size.Width
		height = elem.Size.Height
	}

	// Check page break
	imgHeight := height
	if imgHeight == 0 {
		imgHeight = 100 // Default estimate
	}
	if err := h.CheckPageBreak(imgHeight); err != nil {
		return err
	}

	// Add image
	if width > 0 && height > 0 {
		h.pdf.Image(imagePath, x, y, &gopdf.Rect{W: width, H: height})
	} else if width > 0 {
		h.pdf.Image(imagePath, x, y, &gopdf.Rect{W: width, H: width})
	} else {
		h.pdf.Image(imagePath, x, y, nil)
	}

	// Update cursor
	if elem.Position == nil && height > 0 {
		h.cursorY = y + height
	}

	return nil
}

// handleTable renders a table element
func (h *Handler) handleTable(elem parser.Element) error {
	if len(elem.Columns) == 0 {
		return fmt.Errorf("table has no columns")
	}

	font := h.getFontConfig(elem.Font)
	cellPadding := &parser.Padding{Top: 5, Bottom: 5, Left: 5, Right: 5}
	if elem.CellPadding != nil {
		cellPadding = elem.CellPadding
	}

	// Calculate column widths
	totalWidth := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Size != nil && elem.Size.Width > 0 {
		totalWidth = elem.Size.Width
	}

	colWidths := make([]float64, len(elem.Columns))
	for i, col := range elem.Columns {
		if col.Width > 0 {
			colWidths[i] = col.Width
		} else {
			colWidths[i] = totalWidth / float64(len(elem.Columns))
		}
	}

	// Get starting position
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	// Render header
	if elem.Header != nil {
		headerHeight := h.calculateRowHeight(elem.Header.Cells, colWidths, cellPadding, font)
		if err := h.CheckPageBreak(headerHeight); err != nil {
			return err
		}

		if err := h.renderTableRow(x, y, elem.Header.Cells, colWidths, headerHeight, cellPadding, elem.Header.Font, elem.Header.Background, elem.Border, elem.BorderColor); err != nil {
			return err
		}
		y += headerHeight
	}

	// Render rows
	for _, row := range elem.Rows {
		rowHeight := h.calculateRowHeight(row.Cells, colWidths, cellPadding, font)
		if err := h.CheckPageBreak(rowHeight); err != nil {
			return err
		}

		if err := h.renderTableRow(x, y, row.Cells, colWidths, rowHeight, cellPadding, font, nil, elem.Border, elem.BorderColor); err != nil {
			return err
		}
		y += rowHeight
	}

	// Update cursor
	if elem.Position == nil {
		h.cursorY = y
	}

	return nil
}

// renderTableRow renders a single table row
func (h *Handler) renderTableRow(x, y float64, cells []parser.TableCell, colWidths []float64, rowHeight float64, padding *parser.Padding, font *parser.FontConfig, bgColor *parser.Color, border *parser.Border, borderColor *parser.Color) error {
	cellX := x

	for i, cell := range cells {
		if i >= len(colWidths) {
			break
		}

		cellWidth := colWidths[i]

		// Handle colspan
		if cell.ColSpan > 1 {
			for j := i + 1; j < i+cell.ColSpan && j < len(colWidths); j++ {
				cellWidth += colWidths[j]
			}
		}

		// Draw background
		if cell.Background != nil {
			h.pdf.SetFillColor(cell.Background.R, cell.Background.G, cell.Background.B)
			h.pdf.RectFromUpperLeftWithStyle(cellX, y, cellWidth, rowHeight, "F")
		} else if bgColor != nil {
			h.pdf.SetFillColor(bgColor.R, bgColor.G, bgColor.B)
			h.pdf.RectFromUpperLeftWithStyle(cellX, y, cellWidth, rowHeight, "F")
		}

		// Draw border
		if border != nil {
			h.drawBorder(cellX, y, cellWidth, rowHeight, border, borderColor)
		}

		// Render cell text
		cellFont := h.getFontConfig(font)
		if cell.Font != nil {
			cellFont = h.getFontConfig(cell.Font)
		}

		style := cellFont.Style
		if err := h.fontMgr.SetFont(h.pdf, cellFont.Family, style, cellFont.Size); err != nil {
			return err
		}
		if cellFont.Color != nil {
			h.pdf.SetTextColor(cellFont.Color.R, cellFont.Color.G, cellFont.Color.B)
		} else {
			h.pdf.SetTextColor(0, 0, 0)
		}

		// Process RTL text
		text := cell.Text
		if cell.RTL || rtl.IsRTLText(text) {
			text = rtl.ProcessRTLText(text)
		}

		// Calculate text position with padding
		textX := cellX + padding.Left
		textY := y + padding.Top
		textWidth := cellWidth - padding.Left - padding.Right
		textHeight := rowHeight - padding.Top - padding.Bottom

		// Apply alignment
		alignStr := "LT"
		if cell.RTL || rtl.IsRTLText(cell.Text) {
			alignStr = "RT"
		}
		if cell.Align != "" {
			switch cell.Align {
			case "C", "center":
				alignStr = "CT"
			case "R", "right":
				alignStr = "RT"
			case "L", "left":
				alignStr = "LT"
			}
		}

		// Convert alignment string to int constant
		align := h.parseAlign(alignStr)

		h.pdf.SetXY(textX, textY)
		h.pdf.CellWithOption(&gopdf.Rect{
			W: textWidth,
			H: textHeight,
		}, text, gopdf.CellOption{
			Align:  align,
			Border: 0,
			Float:  gopdf.Left,
		})

		cellX += cellWidth
	}

	return nil
}

// calculateRowHeight calculates the height needed for a row
func (h *Handler) calculateRowHeight(cells []parser.TableCell, colWidths []float64, padding *parser.Padding, defaultFont *parser.FontConfig) float64 {
	maxHeight := 0.0

	for i, cell := range cells {
		if i >= len(colWidths) {
			break
		}

		font := defaultFont
		if cell.Font != nil {
			font = cell.Font
		}

		if font == nil {
			font = &parser.FontConfig{Size: 12}
		}

		// Estimate height based on text length and column width
		cellWidth := colWidths[i] - padding.Left - padding.Right
		lines := float64(len(cell.Text)) / (cellWidth / (font.Size * 0.6))
		if lines < 1 {
			lines = 1
		}

		height := lines*font.Size*1.2 + padding.Top + padding.Bottom
		if height > maxHeight {
			maxHeight = height
		}
	}

	if defaultFont != nil && maxHeight < defaultFont.Size*2 {
		maxHeight = defaultFont.Size * 2
	} else if maxHeight < 24 { // Default minimum height
		maxHeight = 24
	}

	return maxHeight
}

// handleLine renders a line element
func (h *Handler) handleLine(elem parser.Element) error {
	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	} else {
		h.pdf.SetStrokeColor(0, 0, 0)
	}

	lineWidth := elem.LineWidth
	if lineWidth == 0 {
		lineWidth = 1
	}
	h.pdf.SetLineWidth(lineWidth)

	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	endX := elem.EndX
	endY := elem.EndY
	if endX == 0 {
		endX = h.pageWidth - h.margin.Right
	}
	if endY == 0 {
		endY = y
	}

	h.pdf.Line(x, y, endX, endY)

	if elem.Position == nil {
		h.cursorY = y + lineWidth
	}

	return nil
}

// handleRect renders a rectangle element
func (h *Handler) handleRect(elem parser.Element) error {
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	width := elem.Size.Width
	height := elem.Size.Height
	if height == 0 {
		height = width
	}

	// Check page break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}

	style := "D" // Draw

	// Fill
	if elem.FillColor != nil {
		h.pdf.SetFillColor(elem.FillColor.R, elem.FillColor.G, elem.FillColor.B)
		style = "F" // Fill
		if elem.LineColor != nil {
			style = "DF" // Draw and Fill
		}
	}

	// Stroke
	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	}

	if elem.LineWidth > 0 {
		h.pdf.SetLineWidth(elem.LineWidth)
	}

	h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, style)

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// handleEllipse renders an ellipse element
func (h *Handler) handleEllipse(elem parser.Element) error {
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
	}

	width := elem.Size.Width
	height := elem.Size.Height
	if height == 0 {
		height = width
	}

	// Check page break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}

	if elem.FillColor != nil {
		h.pdf.SetFillColor(elem.FillColor.R, elem.FillColor.G, elem.FillColor.B)
	}

	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	}

	if elem.LineWidth > 0 {
		h.pdf.SetLineWidth(elem.LineWidth)
	}

	// gopdf.Oval uses bounding box
	h.pdf.Oval(x, y, x+width, y+height)

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// handleNewline adds vertical space
func (h *Handler) handleNewline(elem parser.Element) error {
	height := elem.Height
	if height == 0 {
		height = h.defaultFont.Size
	}

	h.cursorY += height
	return nil
}

// handlePageBreak adds a new page
func (h *Handler) handlePageBreak() error {
	h.pdf.AddPage()
	h.cursorY = h.margin.Top
	return nil
}

// drawBorder draws borders around a rectangle
func (h *Handler) drawBorder(x, y, width, height float64, border *parser.Border, color *parser.Color) {
	if border.All {
		border.Top = true
		border.Bottom = true
		border.Left = true
		border.Right = true
	}

	if color != nil {
		h.pdf.SetStrokeColor(color.R, color.G, color.B)
	} else {
		h.pdf.SetStrokeColor(0, 0, 0)
	}

	h.pdf.SetLineWidth(0.5)

	if border.Top {
		h.pdf.Line(x, y, x+width, y)
	}
	if border.Bottom {
		h.pdf.Line(x, y+height, x+width, y+height)
	}
	if border.Left {
		h.pdf.Line(x, y, x, y+height)
	}
	if border.Right {
		h.pdf.Line(x+width, y, x+width, y+height)
	}
}

// wrapText wraps text to fit within width
func (h *Handler) wrapText(text string, maxWidth float64, font *parser.FontConfig) []string {
	if text == "" {
		return []string{""}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		width, _ := h.pdf.MeasureTextWidth(testLine)

		if width <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	lines = append(lines, currentLine)
	return lines
}

// getFontConfig returns font config with defaults
func (h *Handler) getFontConfig(font *parser.FontConfig) *parser.FontConfig {
	if font == nil {
		return h.defaultFont
	}

	result := &parser.FontConfig{
		Family: font.Family,
		Size:   font.Size,
		Style:  font.Style,
		Color:  font.Color,
	}

	if result.Family == "" {
		result.Family = h.defaultFont.Family
	}
	if result.Size == 0 {
		result.Size = h.defaultFont.Size
	}

	return result
}

// parseAlign converts string alignment to gopdf int constant
func (h *Handler) parseAlign(align string) int {
	// Start with Left as default
	res := gopdf.Left

	if strings.Contains(align, "R") {
		res = gopdf.Right
	} else if strings.Contains(align, "C") {
		res = gopdf.Center
	}

	if strings.Contains(align, "T") {
		res |= gopdf.Top
	} else if strings.Contains(align, "B") {
		res |= gopdf.Bottom
	} else if strings.Contains(align, "M") {
		res |= gopdf.Middle
	}

	return res
}
