package elements

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
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
	case "card", "container", "div":
		err = h.handleCard(elem)
	case "arc":
		err = h.handleArc(elem)
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

// handleCard renders a card/container element with child elements
func (h *Handler) handleCard(elem parser.Element) error {
	x, y := h.getPosition(elem)

	width := h.pageWidth - h.margin.Left - h.margin.Right
	height := elem.Height
	if elem.Size != nil {
		if elem.Size.Width > 0 {
			width = elem.Size.Width
		}
		if elem.Size.Height > 0 {
			height = elem.Size.Height
		}
	}

	// Snap sizes if absolute positioned
	if elem.Position != nil {
		width = h.snapSize(width)
		if height > 0 {
			height = h.snapSize(height)
		}
	}

	// Calculate content height if not specified
	if height == 0 && len(elem.Elements) > 0 {
		height = h.calculateCardContentHeight(elem)
	}
	if height == 0 {
		height = 100 // Default height
	}

	// 1. Sync Global
	h.cursorY = y
	// 2. Check Break
	if err := h.CheckPageBreak(height); err != nil {
		return err
	}
	// 3. Sync Local
	y = h.cursorY

	// Apply rotation if specified
	if elem.Rotation != 0 {
		// Save current state, apply rotation, render, restore
		// Note: gopdf doesn't support rotation directly, so we skip for now
	}

	// Draw shadow if specified
	if elem.Shadow != nil {
		h.drawCardShadow(x, y, width, height, elem)
	}

	// Draw card background with optional rounded corners
	if elem.BackgroundColor != nil || elem.FillColor != nil {
		bgColor := elem.BackgroundColor
		if bgColor == nil {
			bgColor = elem.FillColor
		}
		h.pdf.SetFillColor(bgColor.R, bgColor.G, bgColor.B)

		if elem.Radius > 0 || elem.CornerRadius > 0 {
			radius := elem.Radius
			if radius == 0 {
				radius = elem.CornerRadius
			}
			h.drawRoundedRect(x, y, width, height, radius, "F")
		} else {
			h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, "F")
		}
	}

	// Draw card border
	if elem.Border != nil || elem.LineColor != nil {
		border := elem.Border
		if border == nil && elem.LineColor != nil {
			border = &parser.Border{All: true}
		}

		lineWidth := elem.LineWidth
		if lineWidth == 0 {
			lineWidth = 1
		}
		if border.Width > 0 {
			lineWidth = border.Width
		}

		borderColor := elem.BorderColor
		if borderColor == nil {
			borderColor = elem.LineColor
		}
		if borderColor == nil {
			borderColor = &parser.Color{R: 0, G: 0, B: 0}
		}

		h.pdf.SetStrokeColor(borderColor.R, borderColor.G, borderColor.B)
		h.pdf.SetLineWidth(lineWidth)
		h.applyLineStyle(border.Style)

		if elem.Radius > 0 || elem.CornerRadius > 0 {
			radius := elem.Radius
			if radius == 0 {
				radius = elem.CornerRadius
			}
			h.drawRoundedRect(x, y, width, height, radius, "D")
		} else {
			h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, "D")
		}
		h.pdf.SetLineType("")
	}

	// Render child elements within card bounds
	if len(elem.Elements) > 0 {
		// Calculate content area with padding
		padding := elem.Padding
		if padding == nil {
			padding = &parser.Padding{Top: 10, Bottom: 10, Left: 10, Right: 10}
		}

		contentX := x + padding.Left
		contentY := y + padding.Top
		contentWidth := width - padding.Left - padding.Right

		// Save current state to restore after card rendering
		originalMarginLeft := h.margin.Left
		originalMarginRight := h.margin.Right
		originalCursorY := h.cursorY
		originalPageWidth := h.pageWidth

		// Temporarily adjust internal state for child elements
		// This creates a "local coordinate system" for relative positioning
		h.margin.Left = contentX
		h.margin.Right = h.pageWidth - (contentX + contentWidth)
		h.cursorY = contentY
		h.pageWidth = contentX + contentWidth

		// Sort child elements by ZIndex
		sortedElements := make([]parser.Element, len(elem.Elements))
		copy(sortedElements, elem.Elements)
		sort.SliceStable(sortedElements, func(i, j int) bool {
			return sortedElements[i].ZIndex < sortedElements[j].ZIndex
		})

		// Render child elements
		for _, child := range sortedElements {
			// Save child's original position if any, to restore it (avoid mutating template)
			var originalChildPos *parser.Position
			if child.Position != nil {
				posCopy := *child.Position
				originalChildPos = &posCopy

				// If child has absolute position, make it relative to card content area
				child.Position.X += contentX
				child.Position.Y += contentY
			} else {
				// Flow positioning: use current cursorY within card
				child.Position = &parser.Position{X: contentX, Y: h.cursorY}
			}

			if err := h.HandleElement(child); err != nil {
				// Try to restore state before returning error
				h.margin.Left = originalMarginLeft
				h.margin.Right = originalMarginRight
				h.cursorY = originalCursorY
				h.pageWidth = originalPageWidth
				return err
			}

			// Restore child position to avoid template pollution
			if originalChildPos != nil {
				*child.Position = *originalChildPos
			} else {
				child.Position = nil
			}
		}

		// Restore original state
		h.margin.Left = originalMarginLeft
		h.margin.Right = originalMarginRight
		h.cursorY = originalCursorY
		h.pageWidth = originalPageWidth
	}

	// Update cursor to after the card if it was in flow
	if elem.Position == nil {
		h.cursorY = y + height
	}

	return nil
}

// calculateCardContentHeight calculates the total height needed for card content
func (h *Handler) calculateCardContentHeight(elem parser.Element) float64 {
	height := 0.0
	padding := elem.Padding
	if padding == nil {
		padding = &parser.Padding{Top: 10, Bottom: 10, Left: 10, Right: 10}
	}
	height += padding.Top + padding.Bottom

	for _, child := range elem.Elements {
		if child.Size != nil && child.Size.Height > 0 {
			height += child.Size.Height
		} else if child.Height > 0 {
			height += child.Height
		} else if child.Type == "text" || child.Type == "cell" {
			font := h.getFontConfig(child.Font)
			height += font.Size * 1.5
		} else if child.Type == "newline" || child.Type == "br" {
			if child.Height > 0 {
				height += child.Height
			} else {
				height += 12
			}
		}
	}

	return height
}

// drawCardShadow draws a shadow for the card
func (h *Handler) drawCardShadow(x, y, width, height float64, elem parser.Element) {
	shadow := elem.Shadow
	if shadow == nil {
		return
	}

	// Default shadow properties
	offsetX := shadow.OffsetX
	if offsetX == 0 {
		offsetX = 3
	}
	offsetY := shadow.OffsetY
	if offsetY == 0 {
		offsetY = 3
	}

	shadowColor := shadow.Color
	if shadowColor == nil {
		shadowColor = &parser.Color{R: 0, G: 0, B: 0, A: 50}
	}

	// Draw shadow rectangle (simplified - gopdf doesn't support blur)
	opacity := float64(shadowColor.A) / 255.0
	if opacity == 0 {
		opacity = 0.2
	}

	h.pdf.SetFillColor(uint8(float64(shadowColor.R)*opacity),
		uint8(float64(shadowColor.G)*opacity),
		uint8(float64(shadowColor.B)*opacity))

	shadowX := x + offsetX
	shadowY := y + offsetY

	if elem.Radius > 0 || elem.CornerRadius > 0 {
		radius := elem.Radius
		if radius == 0 {
			radius = elem.CornerRadius
		}
		h.drawRoundedRect(shadowX, shadowY, width, height, radius, "F")
	} else {
		h.pdf.RectFromUpperLeftWithStyle(shadowX, shadowY, width, height, "F")
	}
}

// drawRoundedRect draws a rectangle with rounded corners
func (h *Handler) drawRoundedRect(x, y, width, height, radius float64, style string) {
	// Clamp radius to not exceed half of width or height
	maxRadius := math.Min(width/2, height/2)
	if radius > maxRadius {
		radius = maxRadius
	}

	if radius <= 0 {
		h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, style)
		return
	}

	segments := 15
	var points []gopdf.Point

	// Top-left corner
	cx, cy := x+radius, y+radius
	for i := 0; i <= segments; i++ {
		angle := math.Pi + (math.Pi/2)*float64(i)/float64(segments)
		points = append(points, gopdf.Point{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)})
	}

	// Top-right corner
	cx, cy = x+width-radius, y+radius
	for i := 0; i <= segments; i++ {
		angle := -math.Pi/2 + (math.Pi/2)*float64(i)/float64(segments)
		points = append(points, gopdf.Point{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)})
	}

	// Bottom-right corner
	cx, cy = x+width-radius, y+height-radius
	for i := 0; i <= segments; i++ {
		angle := 0 + (math.Pi/2)*float64(i)/float64(segments)
		points = append(points, gopdf.Point{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)})
	}

	// Bottom-left corner
	cx, cy = x+radius, y+height-radius
	for i := 0; i <= segments; i++ {
		angle := math.Pi/2 + (math.Pi/2)*float64(i)/float64(segments)
		points = append(points, gopdf.Point{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)})
	}

	// Close the path
	points = append(points, points[0])

	h.pdf.Polygon(points, style)
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

// handleTable renders a table element with enhanced HTML-table-like features
func (h *Handler) handleTable(elem parser.Element) error {
	if len(elem.Columns) == 0 && len(elem.Rows) == 0 {
		return fmt.Errorf("table has no columns or rows")
	}

	font := h.getFontConfig(elem.Font)
	defaultPadding := &parser.Padding{Top: 5, Bottom: 5, Left: 5, Right: 5}
	if elem.CellPadding != nil {
		defaultPadding = elem.CellPadding
	}

	totalWidth := h.pageWidth - h.margin.Left - h.margin.Right
	if elem.Width > 0 {
		totalWidth = elem.Width
	} else if elem.Size != nil && elem.Size.Width > 0 {
		totalWidth = elem.Size.Width
	}
	if elem.Position != nil {
		totalWidth = h.snapSize(totalWidth)
	}

	colWidths := h.calculateColumnWidths(elem, totalWidth)

	x := h.margin.Left
	y := h.cursorY
	if elem.Position != nil {
		x, y = h.getPosition(elem)
	}

	if elem.Caption != "" {
		captionFont := elem.CaptionStyle
		if captionFont == nil {
			captionFont = font
		}
		if err := h.fontMgr.SetFont(h.pdf, captionFont.Family, captionFont.Style, captionFont.Size); err == nil {
			if captionFont.Color != nil {
				h.pdf.SetTextColor(captionFont.Color.R, captionFont.Color.G, captionFont.Color.B)
			}
			h.pdf.SetXY(x, y)
			h.pdf.Cell(nil, elem.Caption)
			y += captionFont.Size * 1.5
		}
	}

	// 1. Pre-calculate row heights to safely accommodate row spans
	rowHeights := make([]float64, len(elem.Rows))
	heightTracker := make(map[int]map[int]bool)

	for rowIndex, row := range elem.Rows {
		rowHeight := h.calculateRowHeightEnhanced(row, colWidths, defaultPadding, font, rowIndex, heightTracker)
		if elem.MinRowHeight > 0 && rowHeight < elem.MinRowHeight {
			rowHeight = elem.MinRowHeight
		}
		if row.MinHeight > 0 && rowHeight < row.MinHeight {
			rowHeight = row.MinHeight
		}
		if row.Height > 0 {
			rowHeight = row.Height
		}
		if len(elem.RowHeights) > rowIndex && elem.RowHeights[rowIndex] > 0 {
			rowHeight = elem.RowHeights[rowIndex]
		}
		rowHeights[rowIndex] = rowHeight

		colIndex := 0
		for _, cell := range row.Cells {
			for heightTracker[rowIndex] != nil && heightTracker[rowIndex][colIndex] {
				colIndex++
			}
			if colIndex >= len(colWidths) {
				break
			}
			colSpan := cell.ColSpan
			if colSpan < 1 {
				colSpan = 1
			}
			rowSpan := cell.RowSpan
			if rowSpan < 1 {
				rowSpan = 1
			}

			if rowSpan > 1 || colSpan > 1 {
				for r := 0; r < rowSpan; r++ {
					for c := 0; c < colSpan; c++ {
						if r == 0 && c == 0 {
							continue
						}
						if colIndex+c >= len(colWidths) {
							continue
						}
						if heightTracker[rowIndex+r] == nil {
							heightTracker[rowIndex+r] = make(map[int]bool)
						}
						heightTracker[rowIndex+r][colIndex+c] = true
					}
				}
			}
			colIndex += colSpan
		}
	}

	// 2. Render Header
	if elem.Header != nil {
		headerHeight := h.calculateHeaderHeight(elem.Header, colWidths, defaultPadding, font)

		h.cursorY = y
		if err := h.CheckPageBreak(headerHeight); err != nil {
			return err
		}
		y = h.cursorY

		if err := h.renderTableHeader(x, y, elem.Header, colWidths, headerHeight, defaultPadding, elem); err != nil {
			return err
		}
		y += headerHeight
	}

	// 3. Render Rows
	rowSpanTracker := make(map[int]map[int]bool)
	for rowIndex, row := range elem.Rows {
		rowHeight := rowHeights[rowIndex]

		h.cursorY = y
		if err := h.CheckPageBreak(rowHeight); err != nil {
			return err
		}
		y = h.cursorY

		if err := h.renderTableRowEnhanced(x, y, row.Cells, colWidths, rowHeight, rowHeights, rowIndex, defaultPadding, font, row, elem, rowSpanTracker); err != nil {
			return err
		}
		y += rowHeight
	}

	// 4. Render Footer
	if elem.Footer != nil {
		footerHeight := h.calculateHeaderHeight(elem.Footer, colWidths, defaultPadding, font)

		h.cursorY = y
		if err := h.CheckPageBreak(footerHeight); err != nil {
			return err
		}
		y = h.cursorY

		if err := h.renderTableFooter(x, y, elem.Footer, colWidths, footerHeight, defaultPadding, elem); err != nil {
			return err
		}
		y += footerHeight
	}

	if elem.Position == nil {
		h.cursorY = y
	}

	return nil
}

// spannedCell tracks cells that span multiple rows
type spannedCell struct {
	cell     parser.TableCell
	startRow int
	endRow   int
	startY   float64
	height   float64
	colWidth float64
	cellX    float64
}

// calculateColumnWidths calculates optimal column widths
func (h *Handler) calculateColumnWidths(elem parser.Element, totalWidth float64) []float64 {
	colWidths := make([]float64, len(elem.Columns))

	// First pass: use specified widths
	specifiedTotal := 0.0
	unspecifiedCount := 0
	for i, col := range elem.Columns {
		if col.Width > 0 {
			colWidths[i] = col.Width
			specifiedTotal += col.Width
		} else {
			unspecifiedCount++
		}
	}

	// Distribute remaining width among unspecified columns
	if unspecifiedCount > 0 {
		remainingWidth := totalWidth - specifiedTotal
		if remainingWidth < 0 {
			remainingWidth = 0
		}
		autoWidth := remainingWidth / float64(unspecifiedCount)
		for i := range colWidths {
			if colWidths[i] == 0 {
				colWidths[i] = autoWidth
			}
		}
	}

	// If no columns specified, create evenly distributed columns
	if len(elem.Columns) == 0 && len(elem.Rows) > 0 {
		maxCols := 0
		for _, row := range elem.Rows {
			if len(row.Cells) > maxCols {
				maxCols = len(row.Cells)
			}
		}
		if elem.Header != nil && len(elem.Header.Cells) > maxCols {
			maxCols = len(elem.Header.Cells)
		}
		colWidths = make([]float64, maxCols)
		colWidth := totalWidth / float64(maxCols)
		for i := range colWidths {
			colWidths[i] = colWidth
		}
	}

	return colWidths
}

// calculateHeaderHeight calculates height needed for table header/footer
func (h *Handler) calculateHeaderHeight(header *parser.TableSection, colWidths []float64, padding *parser.Padding, defaultFont *parser.FontConfig) float64 {
	if header == nil {
		return 0
	}
	height := header.Height
	if height == 0 {
		height = h.calculateRowHeightEnhanced(parser.TableRow{Cells: header.Cells}, colWidths, padding, defaultFont, 0, nil)
	}
	return height
}

// calculateRowHeightEnhanced calculates row height with row span consideration
func (h *Handler) calculateRowHeightEnhanced(row parser.TableRow, colWidths []float64, padding *parser.Padding, defaultFont *parser.FontConfig, rowIndex int, rowSpanTracker map[int]map[int]bool) float64 {
	maxHeight := 0.0
	colIndex := 0

	for _, cell := range row.Cells {
		for rowSpanTracker != nil && rowSpanTracker[rowIndex] != nil && rowSpanTracker[rowIndex][colIndex] {
			colIndex++
		}
		if colIndex >= len(colWidths) {
			break
		}

		mergedFont := &parser.FontConfig{}
		if defaultFont != nil {
			*mergedFont = *defaultFont
		} else {
			mergedFont.Family = "Helvetica"
			mergedFont.Size = 12
		}
		if cell.Font != nil {
			if cell.Font.Family != "" {
				mergedFont.Family = cell.Font.Family
			}
			if cell.Font.Size > 0 {
				mergedFont.Size = cell.Font.Size
			}
		}

		cellPadding := padding
		if cell.Padding != nil {
			cellPadding = cell.Padding
		}

		colSpan := cell.ColSpan
		if colSpan < 1 {
			colSpan = 1
		}
		cellWidth := 0.0
		for j := 0; j < colSpan && colIndex+j < len(colWidths); j++ {
			cellWidth += colWidths[colIndex+j]
		}
		cellWidth -= (cellPadding.Left + cellPadding.Right)

		lines := 1.0
		if cellWidth > 0 {
			lines = float64(len(cell.Text)) / (cellWidth / (mergedFont.Size * 0.6))
			if lines < 1 {
				lines = 1
			}
		}

		rowSpan := cell.RowSpan
		if rowSpan < 1 {
			rowSpan = 1
		}

		height := (lines*mergedFont.Size*1.2 + cellPadding.Top + cellPadding.Bottom)
		if rowSpan > 1 {
			height = height / float64(rowSpan)
		}

		if height > maxHeight {
			maxHeight = height
		}
		colIndex += colSpan
	}

	if maxHeight < 20 {
		maxHeight = 20
	}

	return maxHeight
}

// renderTableHeader renders table header
func (h *Handler) renderTableHeader(x, y float64, header *parser.TableSection, colWidths []float64, height float64, padding *parser.Padding, elem parser.Element) error {
	if header.Background != nil {
		h.pdf.SetFillColor(header.Background.R, header.Background.G, header.Background.B)
		h.pdf.RectFromUpperLeftWithStyle(x, y, sum(colWidths), height, "F")
	}

	if header.Border != nil || elem.Border != nil {
		border := header.Border
		if border == nil {
			border = elem.Border
		}
		borderColor := header.BorderColor
		if borderColor == nil {
			borderColor = elem.BorderColor
		}
		h.drawBorder(x, y, sum(colWidths), height, border, borderColor)
	}

	return h.renderTableRowEnhanced(x, y, header.Cells, colWidths, height, nil, -1, padding, header.Font, parser.TableRow{}, elem, nil)
}

// renderTableFooter renders table footer
func (h *Handler) renderTableFooter(x, y float64, footer *parser.TableSection, colWidths []float64, height float64, padding *parser.Padding, elem parser.Element) error {
	if footer.Background != nil {
		h.pdf.SetFillColor(footer.Background.R, footer.Background.G, footer.Background.B)
		h.pdf.RectFromUpperLeftWithStyle(x, y, sum(colWidths), height, "F")
	}

	if footer.Border != nil || elem.Border != nil {
		border := footer.Border
		if border == nil {
			border = elem.Border
		}
		borderColor := footer.BorderColor
		if borderColor == nil {
			borderColor = elem.BorderColor
		}
		h.drawBorder(x, y, sum(colWidths), height, border, borderColor)
	}

	return h.renderTableRowEnhanced(x, y, footer.Cells, colWidths, height, nil, -1, padding, footer.Font, parser.TableRow{}, elem, nil)
}

// renderTableRowEnhanced renders a table row with full HTML-table-like support
func (h *Handler) renderTableRowEnhanced(x, y float64, cells []parser.TableCell, colWidths []float64, rowHeight float64, rowHeights []float64, rowIndex int, defaultPadding *parser.Padding, defaultFont *parser.FontConfig, row parser.TableRow, elem parser.Element, rowSpanTracker map[int]map[int]bool) error {
	cellX := x
	colIndex := 0

	for _, cell := range cells {
		// Skip cells that are part of a row span from previous rows
		for rowSpanTracker != nil && rowSpanTracker[rowIndex] != nil && rowSpanTracker[rowIndex][colIndex] {
			cellX += colWidths[colIndex]
			colIndex++
		}

		if colIndex >= len(colWidths) {
			break
		}

		colSpan := cell.ColSpan
		if colSpan < 1 {
			colSpan = 1
		}
		rowSpan := cell.RowSpan
		if rowSpan < 1 {
			rowSpan = 1
		}

		cellWidth := 0.0
		for j := 0; j < colSpan && colIndex+j < len(colWidths); j++ {
			cellWidth += colWidths[colIndex+j]
		}

		totalCellHeight := rowHeight
		if rowSpan > 1 && rowIndex >= 0 && rowHeights != nil {
			for r := 1; r < rowSpan && rowIndex+r < len(rowHeights); r++ {
				totalCellHeight += rowHeights[rowIndex+r]
			}
			if rowSpanTracker != nil {
				for r := 0; r < rowSpan; r++ {
					for c := 0; c < colSpan; c++ {
						if r == 0 && c == 0 {
							continue
						}
						if colIndex+c >= len(colWidths) {
							continue
						}
						if rowSpanTracker[rowIndex+r] == nil {
							rowSpanTracker[rowIndex+r] = make(map[int]bool)
						}
						rowSpanTracker[rowIndex+r][colIndex+c] = true
					}
				}
			}
		}

		cellPadding := defaultPadding
		if cell.Padding != nil {
			cellPadding = cell.Padding
		}

		bgColor := cell.Background
		if bgColor == nil && row.Background != nil {
			bgColor = row.Background
		}
		if bgColor != nil {
			h.pdf.SetFillColor(bgColor.R, bgColor.G, bgColor.B)
			h.pdf.RectFromUpperLeftWithStyle(cellX, y, cellWidth, totalCellHeight, "F")
		}

		if cell.Border != nil {
			h.drawBorder(cellX, y, cellWidth, totalCellHeight, cell.Border, cell.BorderColor)
		} else if row.Border != nil {
			h.drawBorder(cellX, y, cellWidth, totalCellHeight, row.Border, row.BorderColor)
		} else if elem.Border != nil {
			h.drawBorder(cellX, y, cellWidth, totalCellHeight, elem.Border, elem.BorderColor)
		}

		if len(cell.Elements) > 0 {
			for _, nestedElem := range cell.Elements {
				if nestedElem.Position == nil {
					nestedElem.Position = &parser.Position{
						X: cellX + cellPadding.Left,
						Y: y + cellPadding.Top,
					}
				}
				if err := h.HandleElement(nestedElem); err != nil {
					return err
				}
			}
		} else if cell.Text != "" {
			mergedFont := &parser.FontConfig{}
			if defaultFont != nil {
				*mergedFont = *defaultFont
			} else {
				mergedFont.Family = "Helvetica"
				mergedFont.Size = 12
			}

			// Ensure we properly merge the Font style logic without abandoning the default family if omitted
			if cell.Font != nil {
				if cell.Font.Family != "" {
					mergedFont.Family = cell.Font.Family
				}
				if cell.Font.Size > 0 {
					mergedFont.Size = cell.Font.Size
				}
				if cell.Font.Style != "" {
					mergedFont.Style = cell.Font.Style
				}
				if cell.Font.Color != nil {
					mergedFont.Color = cell.Font.Color
				}
			}

			if err := h.fontMgr.SetFont(h.pdf, mergedFont.Family, mergedFont.Style, mergedFont.Size); err != nil {
				return err
			}

			if mergedFont.Color != nil {
				h.pdf.SetTextColor(mergedFont.Color.R, mergedFont.Color.G, mergedFont.Color.B)
			} else {
				h.pdf.SetTextColor(0, 0, 0)
			}

			text := cell.Text
			if cell.RTL || rtl.IsRTLText(text) {
				text = rtl.ProcessRTLText(text)
			}

			textX := cellX + cellPadding.Left
			textY := y + cellPadding.Top
			textWidth := cellWidth - cellPadding.Left - cellPadding.Right
			textHeight := totalCellHeight - cellPadding.Top - cellPadding.Bottom

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
			if cell.VerticalAlign != "" {
				switch cell.VerticalAlign {
				case "M", "middle":
					alignStr = alignStr[:1] + "M"
				case "B", "bottom":
					alignStr = alignStr[:1] + "B"
				case "T", "top":
					alignStr = alignStr[:1] + "T"
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
		}

		cellX += cellWidth
		colIndex += colSpan
	}

	return nil
}

// Helper function to sum slice
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

// Legacy renderTableRow for backward compatibility
func (h *Handler) renderTableRow(x, y float64, cells []parser.TableCell, colWidths []float64, rowHeight float64, padding *parser.Padding, font *parser.FontConfig, bgColor *parser.Color, border *parser.Border, borderColor *parser.Color) error {
	return h.renderTableRowEnhanced(x, y, cells, colWidths, rowHeight, nil, -1, padding, font, parser.TableRow{Background: bgColor}, parser.Element{Border: border, BorderColor: borderColor}, nil)
}

// Legacy calculateRowHeight for backward compatibility
func (h *Handler) calculateRowHeight(cells []parser.TableCell, colWidths []float64, padding *parser.Padding, defaultFont *parser.FontConfig) float64 {
	return h.calculateRowHeightEnhanced(parser.TableRow{Cells: cells}, colWidths, padding, defaultFont, 0, nil)
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

	// Draw rounded rectangle if radius specified
	if elem.CornerRadius > 0 || elem.Radius > 0 {
		radius := elem.CornerRadius
		if radius == 0 {
			radius = elem.Radius
		}
		h.drawRoundedRect(x, y, width, height, radius, style)
	} else {
		h.pdf.RectFromUpperLeftWithStyle(x, y, width, height, style)
	}
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

// handleArc renders an arc or pie segment
func (h *Handler) handleArc(elem parser.Element) error {
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

	// Set colors
	if elem.FillColor != nil {
		h.pdf.SetFillColor(elem.FillColor.R, elem.FillColor.G, elem.FillColor.B)
	}
	if elem.LineColor != nil {
		h.pdf.SetStrokeColor(elem.LineColor.R, elem.LineColor.G, elem.LineColor.B)
	}
	if elem.LineWidth > 0 {
		h.pdf.SetLineWidth(elem.LineWidth)
	}

	// Draw ellipse (gopdf doesn't support arcs natively, so we draw full ellipse)
	// For true arc support, would need to use path commands
	h.pdf.Oval(x, y, x+width, y+height)

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

	lineWidth := 0.5
	if border.Width > 0 {
		lineWidth = border.Width
	}
	h.pdf.SetLineWidth(lineWidth)
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
