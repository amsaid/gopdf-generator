package elements

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
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
	pdf          *gopdf.GoPdf
	fontMgr      *fonts.Manager
	cursorY      float64
	pageWidth    float64
	pageHeight   float64
	margin       *parser.Margin
	defaultFont  *parser.FontConfig
	tmpl         *parser.DocumentTemplate
	inDecoration bool
}

// NewHandler creates a new element handler
func NewHandler(pdf *gopdf.GoPdf, fontMgr *fonts.Manager, pageWidth, pageHeight float64, tmpl *parser.DocumentTemplate) *Handler {
	return &Handler{
		pdf:          pdf,
		fontMgr:      fontMgr,
		cursorY:      tmpl.Margin.Top,
		pageWidth:    pageWidth,
		pageHeight:   pageHeight,
		margin:       tmpl.Margin,
		defaultFont:  tmpl.DefaultFont,
		tmpl:         tmpl,
		inDecoration: false,
	}
}

// DrawPageDecorations draws headers, footers, and page backgrounds
func (h *Handler) DrawPageDecorations() error {
	h.inDecoration = true
	defer func() { h.inDecoration = false }()

	originalY := h.cursorY

	// Draw Page Background if defined
	if h.tmpl.Background != nil {
		h.pdf.SetFillColor(h.tmpl.Background.R, h.tmpl.Background.G, h.tmpl.Background.B)
		h.pdf.RectFromUpperLeftWithStyle(0, 0, h.pageWidth, h.pageHeight, "F")
	}

	// Draw Grid if requested
	if h.tmpl.Grid != nil && h.tmpl.Grid.Draw && h.tmpl.Grid.Size > 0 {
		gridSize := h.tmpl.Grid.Size
		if gridSize < 5 {
			gridSize = 5 // Enforce a minimum grid size
		}
		h.pdf.SetStrokeColor(220, 220, 220)
		h.pdf.SetLineWidth(0.5)
		h.pdf.SetLineType("dashed")

		for x := 0.0; x < h.pageWidth; x += gridSize {
			h.pdf.Line(x, 0, x, h.pageHeight)
		}
		for y := 0.0; y < h.pageHeight; y += gridSize {
			h.pdf.Line(0, y, h.pageWidth, y)
		}
		h.pdf.SetLineType("") // reset
	}

	// Draw Watermark
	if h.tmpl.Watermark != nil && h.tmpl.Watermark.Text != "" {
		font := h.getFontConfig(h.tmpl.Watermark.Font)
		style := font.Style
		if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err == nil {
			if font.Color != nil {
				h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
			} else {
				h.pdf.SetTextColor(200, 200, 200)
			}

			w, _ := h.pdf.MeasureTextWidth(h.tmpl.Watermark.Text)
			x := (h.pageWidth - w) / 2
			y := h.pageHeight / 2

			h.pdf.SetXY(x, y)
			h.pdf.Cell(nil, h.tmpl.Watermark.Text)
		}
	}

	// Draw Header Elements
	if len(h.tmpl.Header) > 0 {
		h.cursorY = h.margin.Top / 2
		for _, elem := range h.tmpl.Header {
			if err := h.HandleElement(elem); err != nil {
				return err
			}
		}
	}

	// Draw Footer Elements
	if len(h.tmpl.Footer) > 0 {
		h.cursorY = h.pageHeight - (h.margin.Bottom * 0.75)
		for _, elem := range h.tmpl.Footer {
			if err := h.HandleElement(elem); err != nil {
				return err
			}
		}
	}

	h.cursorY = originalY
	return nil
}

// SetCursorY sets the current Y position
func (h *Handler) SetCursorY(y float64) {
	h.cursorY = y
}

// GetCursorY returns the current Y position
func (h *Handler) GetCursorY() float64 {
	return h.cursorY
}

// getPosition returns the resolved X, Y considering grid snap using Nearest Rounding
func (h *Handler) getPosition(elem parser.Element) (float64, float64) {
	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x = elem.Position.X
		y = elem.Position.Y
		if h.tmpl.Grid != nil && h.tmpl.Grid.Size > 0 {
			gridSize := h.tmpl.Grid.Size
			if gridSize < 5 {
				gridSize = 5 // Safe minimum boundary for grid snapping
			}
			// Snaps coordinates to the nearest grid line instead of rounding down
			x = math.Round(x/gridSize) * gridSize
			y = math.Round(y/gridSize) * gridSize
		}
	}
	return x, y
}

// snapSize is a helper to snap sizes (Width/Height) to the grid
func (h *Handler) snapSize(val float64) float64 {
	if h.tmpl.Grid != nil && h.tmpl.Grid.Size > 0 {
		gridSize := h.tmpl.Grid.Size
		if gridSize < 5 {
			gridSize = 5
		}
		return math.Round(val/gridSize) * gridSize
	}
	return val
}

// CheckPageBreak checks if we need a page break
func (h *Handler) CheckPageBreak(requiredHeight float64) error {
	if h.inDecoration {
		return nil
	}
	if h.cursorY+requiredHeight > h.pageHeight-h.margin.Bottom {
		h.pdf.AddPage()
		h.cursorY = h.margin.Top
		return h.DrawPageDecorations()
	}
	return nil
}

// HandleElement processes a single element
func (h *Handler) HandleElement(elem parser.Element) error {
	var err error
	switch elem.Type {
	case "text":
		err = h.handleText(elem)
	case "cell":
		err = h.handleCell(elem)
	case "image":
		err = h.handleImage(elem)
	case "table":
		err = h.handleTable(elem)
	case "line":
		err = h.handleLine(elem)
	case "rect", "rectangle":
		err = h.handleRect(elem)
	case "ellipse":
		err = h.handleEllipse(elem)
	case "newline", "br":
		err = h.handleNewline(elem)
	case "pagebreak":
		err = h.handlePageBreak()
	case "list":
		err = h.handleList(elem)
	case "link":
		err = h.handleLink(elem)
	default:
		err = fmt.Errorf("unknown element type: %s", elem.Type)
	}

	return err
}

func (h *Handler) applyLineStyle(style string) {
	switch style {
	case "dashed":
		h.pdf.SetLineType("dashed")
	case "dotted":
		h.pdf.SetLineType("dotted")
	default:
		h.pdf.SetLineType("")
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

	// Process RTL text
	text := elem.Text
	isRTL := elem.RTL || rtl.IsRTLText(text)
	if isRTL || rtl.ContainsRTL(text) {
		text = rtl.ShapeArabic(text)
	}

	// Get position and apply indent
	x, y := h.getPosition(elem)
	x += elem.Indent

	// Calculate text dimensions
	width := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Size != nil && elem.Size.Width > 0 {
		width = elem.Size.Width
		// Snap width if absolute positioned
		if elem.Position != nil {
			width = h.snapSize(width)
		}
	}
	width -= elem.Indent
	if width <= 0 {
		width = 10 // safe fallback
	}

	// Handle alignment
	align := "L"
	if isRTL {
		align = "R"
	}
	if elem.Alignment != nil && elem.Alignment.Horizontal != "" {
		align = elem.Alignment.Horizontal
	}

	// Calculate line height
	lineHeight := font.Size * 1.2
	if elem.LineHeight > 0 {
		lineHeight = elem.LineHeight
	}

	lines := h.wrapText(text, width, font)

	// Render each line individually to handle page breaks
	for _, line := range lines {
		// 1. Sync global
		h.cursorY = y
		// 2. Check break per line
		if err := h.CheckPageBreak(lineHeight); err != nil {
			return err
		}
		// 3. Sync local
		y = h.cursorY

		lineX := x

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

	x, y := h.getPosition(elem)
	x += elem.Indent

	width := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Size != nil && elem.Size.Width > 0 {
		width = elem.Size.Width
		// Snap width if absolute positioned
		if elem.Position != nil {
			width = h.snapSize(width)
		}
	}
	width -= elem.Indent
	bulletIndent := font.Size * 1.5 // Space reserved for the bullet point

	width -= bulletIndent
	if width <= 0 {
		width = 10
	}

	lineHeight := font.Size * 1.2
	if elem.LineHeight > 0 {
		lineHeight = elem.LineHeight
	}

	for i, item := range elem.ListItems {
		bullet := "•"
		if elem.ListType == "ol" {
			bullet = fmt.Sprintf("%d.", i+1)
		}

		isRTL := elem.RTL || rtl.IsRTLText(item)
		renderItem := item
		if isRTL || rtl.ContainsRTL(item) {
			renderItem = rtl.ShapeArabic(item)
		}

		lines := h.wrapText(renderItem, width, font)
		itemHeight := lineHeight * float64(len(lines))

		// 1. Sync global
		h.cursorY = y
		// 2. Check Break
		if err := h.CheckPageBreak(itemHeight); err != nil {
			return err
		}
		// 3. Sync local
		y = h.cursorY

		bulletX := x
		if isRTL {
			bulletX = x + width + 5
		}
		h.pdf.SetXY(bulletX, y)
		h.pdf.Cell(nil, bullet)

		for _, line := range lines {
			renderLine := line
			if isRTL || rtl.ContainsRTL(line) {
				renderLine = rtl.ReorderString(line, isRTL)
			}

			lineX := x + bulletIndent
			if isRTL {
				lineWidth, _ := h.pdf.MeasureTextWidth(renderLine)
				lineX = x + width - lineWidth
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

	style := font.Style
	if !strings.Contains(strings.ToUpper(style), "U") {
		style += "U"
	}

	if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

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

	x, y := h.getPosition(elem)

	renderText := text
	if isRTL || rtl.ContainsRTL(text) {
		renderText = rtl.ReorderString(text, isRTL)
	}

	w, _ := h.pdf.MeasureTextWidth(renderText)
	height := font.Size * 1.5

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	h.pdf.SetXY(x, y)
	h.pdf.Cell(nil, renderText)
	h.pdf.AddExternalLink(elem.URL, x, y, w, font.Size)

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// handleCell renders a cell element
func (h *Handler) handleCell(elem parser.Element) error {
	font := h.getFontConfig(elem.Font)

	style := font.Style
	if err := h.fontMgr.SetFont(h.pdf, font.Family, style, font.Size); err != nil {
		return fmt.Errorf("setting font %s: %w", font.Family, err)
	}

	if font.Color != nil {
		h.pdf.SetTextColor(font.Color.R, font.Color.G, font.Color.B)
	} else {
		h.pdf.SetTextColor(0, 0, 0)
	}

	x, y := h.getPosition(elem)

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

	// Snap width and height if absolute positioned
	if elem.Position != nil {
		width = h.snapSize(width)
		height = h.snapSize(height)
	}

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	// Draw background
	if elem.BackgroundColor != nil {
		h.pdf.SetFillColor(elem.BackgroundColor.R, elem.BackgroundColor.G, elem.BackgroundColor.B)
		h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, "F")
	}

	// Draw border
	if elem.Border != nil {
		h.drawBorder(x, y, width, height, elem.Border, elem.BorderColor)
	}

	text := elem.Text
	if elem.RTL || rtl.IsRTLText(text) {
		text = rtl.ProcessRTLText(text)
	}

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

	align := h.parseAlign(alignStr)

	h.pdf.SetXY(x, y)
	h.pdf.CellWithOption(&gopdf.Rect{
		W: width,
		H: height,
	}, text, gopdf.CellOption{
		Align:  align,
		Border: 0,
		Float:  gopdf.Left,
	})

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// handleImage renders an image element
func (h *Handler) handleImage(elem parser.Element) error {
	var imagePath string
	var cleanup bool

	if strings.HasSuffix(strings.ToLower(elem.ImagePath), ".svg") || strings.HasPrefix(elem.ImageURL, "data:image/svg") {
		return fmt.Errorf("SVG images are not natively supported by the PDF engine. Please convert to PNG or JPEG")
	}

	if elem.ImagePath != "" {
		imagePath = elem.ImagePath
	} else if len(elem.ImageData) > 0 {
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

	if cleanup {
		defer os.Remove(imagePath)
	}

	x, y := h.getPosition(elem)

	width := 0.0
	height := 0.0
	if elem.Size != nil {
		width = elem.Size.Width
		height = elem.Size.Height

		// Snap sizes if absolute positioned
		if elem.Position != nil {
			if width > 0 {
				width = h.snapSize(width)
			}
			if height > 0 {
				height = h.snapSize(height)
			}
		}
	}

	imgHeight := height
	if imgHeight == 0 {
		imgHeight = 100 // Default estimate
	}

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(imgHeight); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	if width > 0 && height > 0 {
		h.pdf.Image(imagePath, x, y, &gopdf.Rect{W: width, H: height})
	} else if width > 0 {
		h.pdf.Image(imagePath, x, y, &gopdf.Rect{W: width, H: width})
	} else {
		h.pdf.Image(imagePath, x, y, nil)
	}

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
		// Snap table width if absolute positioned
		if elem.Position != nil {
			totalWidth = h.snapSize(totalWidth)
		}
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
		x, y = h.getPosition(elem)
	}

	// Render header
	if elem.Header != nil {
		headerHeight := h.calculateRowHeight(elem.Header.Cells, colWidths, cellPadding, font)

		// 1. Sync global cursor so CheckPageBreak checks the correct location
		h.cursorY = y

		if err := h.CheckPageBreak(headerHeight); err != nil {
			return err
		}

		// 2. Sync local 'y' back (in case a new page was added, y is now h.margin.Top)
		y = h.cursorY

		if err := h.renderTableRow(x, y, elem.Header.Cells, colWidths, headerHeight, cellPadding, elem.Header.Font, elem.Header.Background, elem.Border, elem.BorderColor); err != nil {
			return err
		}
		y += headerHeight
	}

	// Render rows
	for _, row := range elem.Rows {
		rowHeight := h.calculateRowHeight(row.Cells, colWidths, cellPadding, font)

		// 1. Sync global cursor so CheckPageBreak checks the correct location
		h.cursorY = y

		if err := h.CheckPageBreak(rowHeight); err != nil {
			return err
		}

		// 2. Sync local 'y' back (if page break happened, this moves y to the top of the next page)
		y = h.cursorY

		if err := h.renderTableRow(x, y, row.Cells, colWidths, rowHeight, cellPadding, font, nil, elem.Border, elem.BorderColor); err != nil {
			return err
		}
		y += rowHeight
	}

	// Update cursor final position
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

		text := cell.Text
		if cell.RTL || rtl.IsRTLText(text) {
			text = rtl.ProcessRTLText(text)
		}

		textX := cellX + padding.Left
		textY := y + padding.Top
		textWidth := cellWidth - padding.Left - padding.Right
		textHeight := rowHeight - padding.Top - padding.Bottom

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
	} else if maxHeight < 24 {
		maxHeight = 24
	}

	return maxHeight
}

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
	h.applyLineStyle(elem.LineStyle)

	x, y := h.getPosition(elem)

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(lineWidth); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	endX := elem.EndX
	endY := elem.EndY
	if endX == 0 {
		endX = h.pageWidth - h.margin.Right
	}
	if endY == 0 {
		endY = y
	} else if elem.Position == nil && elem.EndY == 0 {
		// If flow positioning, endY is same as y (horizontal line)
		endY = y
	}

	// Snap end coordinates if absolutely positioned
	if elem.Position != nil {
		endX = h.snapSize(endX)
		endY = h.snapSize(endY)
	}

	h.pdf.Line(x, y, endX, endY)
	h.pdf.SetLineType("")

	if elem.Position == nil {
		h.cursorY = y + lineWidth
	}

	return nil
}

func (h *Handler) handleRect(elem parser.Element) error {
	x, y := h.getPosition(elem)

	width := elem.Size.Width
	height := elem.Size.Height
	if height == 0 {
		height = width
	}

	// Snap sizes if absolutely positioned
	if elem.Position != nil {
		width = h.snapSize(width)
		height = h.snapSize(height)
	}

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	style := "D" // Draw

	if elem.FillColor != nil {
		h.pdf.SetFillColor(elem.FillColor.R, elem.FillColor.G, elem.FillColor.B)
		style = "F" // Fill
		if elem.LineColor != nil {
			style = "DF" // Draw and Fill
		}
	}

	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	}

	if elem.LineWidth > 0 {
		h.pdf.SetLineWidth(elem.LineWidth)
	}

	h.applyLineStyle(elem.LineStyle)
	h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, style)
	h.pdf.SetLineType("") // reset

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

func (h *Handler) handleEllipse(elem parser.Element) error {
	x, y := h.getPosition(elem)

	width := elem.Size.Width
	height := elem.Size.Height
	if height == 0 {
		height = width
	}

	// Snap sizes if absolutely positioned
	if elem.Position != nil {
		width = h.snapSize(width)
		height = h.snapSize(height)
	}

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	if elem.FillColor != nil {
		h.pdf.SetFillColor(elem.FillColor.R, elem.FillColor.G, elem.FillColor.B)
	}

	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	}

	if elem.LineWidth > 0 {
		h.pdf.SetLineWidth(elem.LineWidth)
	}

	h.applyLineStyle(elem.LineStyle)
	h.pdf.Oval(x, y, x+width, y+height)
	h.pdf.SetLineType("") // reset

	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

func (h *Handler) handleNewline(elem parser.Element) error {
	height := elem.Height
	if height == 0 {
		if h.defaultFont != nil && h.defaultFont.Size > 0 {
			height = h.defaultFont.Size
		} else {
			height = 12
		}
	}

	if err := h.CheckPageBreak(height); err != nil {
		return err
	}

	h.cursorY += height
	return nil
}

func (h *Handler) handlePageBreak() error {
	h.pdf.AddPage()
	h.cursorY = h.margin.Top
	return h.DrawPageDecorations()
}

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
	h.applyLineStyle(border.Style)

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

	h.pdf.SetLineType("")
}

func (h *Handler) wrapText(text string, maxWidth float64, font *parser.FontConfig) []string {
	if text == "" {
		return []string{""}
	}

	var finalLines []string

	// Split by explicit newlines first to preserve user formatting
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if para == "" {
			// Preserve empty lines
			finalLines = append(finalLines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			finalLines = append(finalLines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			testLine := currentLine + " " + word
			width, _ := h.pdf.MeasureTextWidth(testLine)

			if width <= maxWidth {
				currentLine = testLine
			} else {
				finalLines = append(finalLines, currentLine)
				currentLine = word
			}
		}
		finalLines = append(finalLines, currentLine)
	}

	return finalLines
}

func (h *Handler) getFontConfig(font *parser.FontConfig) *parser.FontConfig {
	var result *parser.FontConfig
	if font == nil {
		if h.defaultFont != nil {
			result = &parser.FontConfig{
				Family: h.defaultFont.Family,
				Size:   h.defaultFont.Size,
				Style:  h.defaultFont.Style,
				Color:  h.defaultFont.Color,
			}
		} else {
			result = &parser.FontConfig{}
		}
	} else {
		result = &parser.FontConfig{
			Family: font.Family,
			Size:   font.Size,
			Style:  font.Style,
			Color:  font.Color,
		}
	}

	if result.Family == "" {
		if h.defaultFont != nil && h.defaultFont.Family != "" {
			result.Family = h.defaultFont.Family
		} else {
			result.Family = "Helvetica"
		}
	}
	if result.Size == 0 {
		if h.defaultFont != nil && h.defaultFont.Size > 0 {
			result.Size = h.defaultFont.Size
		} else {
			result.Size = 12
		}
	}

	return result
}

func (h *Handler) parseAlign(align string) int {
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
