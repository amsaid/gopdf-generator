package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

// DocumentTemplate represents the root structure of a PDF template
type DocumentTemplate struct {
	// Page settings
	PageSize    string  `json:"page_size,omitempty"`   // A4, A3, Letter, Legal, or custom
	PageWidth   float64 `json:"page_width,omitempty"`  // Custom width in points (if PageSize not set)
	PageHeight  float64 `json:"page_height,omitempty"` // Custom height in points
	Orientation string  `json:"orientation,omitempty"` // portrait or landscape
	Margin      *Margin `json:"margin,omitempty"`

	// Document metadata
	Title   string `json:"title,omitempty"`
	Author  string `json:"author,omitempty"`
	Subject string `json:"subject,omitempty"`
	Creator string `json:"creator,omitempty"`

	// Default font settings
	DefaultFont *FontConfig `json:"default_font,omitempty"`

	// Content elements
	Elements []Element `json:"elements"`

	// Font registrations
	Fonts []FontDef `json:"fonts,omitempty"`
}

// Margin represents page margins
type Margin struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// FontConfig represents font configuration
type FontConfig struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Style  string  `json:"style,omitempty"` // B, I, U, or combinations
	Color  *Color  `json:"color,omitempty"`
}

// FontDef represents a font to be registered
type FontDef struct {
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	// For embedded fonts from bytes
	Data []byte `json:"-"`
}

// Color represents RGB color
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// Position represents element position
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents element dimensions
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height,omitempty"`
}

// Border represents cell/table borders
type Border struct {
	Top    bool `json:"top,omitempty"`
	Bottom bool `json:"bottom,omitempty"`
	Left   bool `json:"left,omitempty"`
	Right  bool `json:"right,omitempty"`
	All    bool `json:"all,omitempty"`
}

// Alignment represents text alignment
type Alignment struct {
	Horizontal string `json:"horizontal,omitempty"` // L, C, R, J (left, center, right, justify)
	Vertical   string `json:"vertical,omitempty"`   // T, M, B (top, middle, bottom)
}

// Element represents a generic PDF element
type Element struct {
	// Element type: text, image, table, line, rect, ellipse, cell, newline, pagebreak, list, link
	Type string `json:"type"`

	// Common properties
	Position *Position `json:"position,omitempty"`
	Size     *Size     `json:"size,omitempty"`

	// Text properties
	Text       string      `json:"text,omitempty"`
	Font       *FontConfig `json:"font,omitempty"`
	Alignment  *Alignment  `json:"alignment,omitempty"`
	LineHeight float64     `json:"line_height,omitempty"`

	// RTL support
	RTL bool `json:"rtl,omitempty"`

	// Image properties
	ImagePath string `json:"image_path,omitempty"`
	ImageData []byte `json:"-"`
	ImageURL  string `json:"image_url,omitempty"`

	// Shape properties
	FillColor *Color  `json:"fill_color,omitempty"`
	LineColor *Color  `json:"line_color,omitempty"`
	LineWidth float64 `json:"line_width,omitempty"`

	// Table properties
	Columns     []TableColumn `json:"columns,omitempty"`
	Rows        []TableRow    `json:"rows,omitempty"`
	Header      *TableHeader  `json:"header,omitempty"`
	CellPadding *Padding      `json:"cell_padding,omitempty"`

	// Border properties
	Border      *Border `json:"border,omitempty"`
	BorderColor *Color  `json:"border_color,omitempty"`

	// Cell properties
	BackgroundColor *Color `json:"background_color,omitempty"`

	// Line properties
	EndX float64 `json:"end_x,omitempty"`
	EndY float64 `json:"end_y,omitempty"`

	// List properties
	ListItems []string `json:"list_items,omitempty"`
	ListType  string   `json:"list_type,omitempty"` // "ul" or "ol"

	// Link properties
	URL string `json:"url,omitempty"`

	// Spacing
	Height float64 `json:"height,omitempty"` // For newline
}

// TableColumn represents a table column definition
type TableColumn struct {
	Width float64 `json:"width"`
	Align string  `json:"align,omitempty"` // L, C, R
}

// TableRow represents a table row
type TableRow struct {
	Cells []TableCell `json:"cells"`
}

// TableCell represents a table cell
type TableCell struct {
	Text       string      `json:"text"`
	Font       *FontConfig `json:"font,omitempty"`
	Align      string      `json:"align,omitempty"` // L, C, R
	ColSpan    int         `json:"col_span,omitempty"`
	RowSpan    int         `json:"row_span,omitempty"`
	Background *Color      `json:"background,omitempty"`
	RTL        bool        `json:"rtl,omitempty"`
}

// TableHeader represents table header row
type TableHeader struct {
	Cells      []TableCell `json:"cells"`
	Font       *FontConfig `json:"font,omitempty"`
	Background *Color      `json:"background,omitempty"`
	Repeat     bool        `json:"repeat,omitempty"` // Repeat on each page
}

// Padding represents cell padding
type Padding struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// ParseTemplate parses a JSON template from reader
func ParseTemplate(r io.Reader) (*DocumentTemplate, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}

	var tmpl DocumentTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parsing template JSON: %w", err)
	}

	// Set defaults
	if tmpl.PageSize == "" && tmpl.PageWidth == 0 {
		tmpl.PageSize = "A4"
	}
	if tmpl.Orientation == "" {
		tmpl.Orientation = "portrait"
	}
	if tmpl.Margin == nil {
		tmpl.Margin = &Margin{Top: 72, Bottom: 72, Left: 72, Right: 72}
	}
	if tmpl.DefaultFont == nil {
		tmpl.DefaultFont = &FontConfig{Family: "Helvetica", Size: 12}
	}

	return &tmpl, nil
}

// ParseTemplateBytes parses a JSON template from bytes
func ParseTemplateBytes(data []byte) (*DocumentTemplate, error) {
	var tmpl DocumentTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parsing template JSON: %w", err)
	}

	// Set defaults
	if tmpl.PageSize == "" && tmpl.PageWidth == 0 {
		tmpl.PageSize = "A4"
	}
	if tmpl.Orientation == "" {
		tmpl.Orientation = "portrait"
	}
	if tmpl.Margin == nil {
		tmpl.Margin = &Margin{Top: 72, Bottom: 72, Left: 72, Right: 72}
	}
	if tmpl.DefaultFont == nil {
		tmpl.DefaultFont = &FontConfig{Family: "Helvetica", Size: 12}
	}

	return &tmpl, nil
}

// Validate validates the template
func (t *DocumentTemplate) Validate() error {
	if len(t.Elements) == 0 {
		return fmt.Errorf("template must have at least one element")
	}

	for i, elem := range t.Elements {
		if elem.Type == "" {
			return fmt.Errorf("element %d: type is required", i)
		}

		switch elem.Type {
		case "text", "cell":
			if elem.Text == "" && len(elem.Type) > 0 {
				// Allow empty text
			}
		case "image":
			if elem.ImagePath == "" && len(elem.ImageData) == 0 && elem.ImageURL == "" {
				return fmt.Errorf("element %d: image requires path, data, or URL", i)
			}
		case "table":
			if len(elem.Columns) == 0 {
				return fmt.Errorf("element %d: table requires columns", i)
			}
		case "list":
			if len(elem.ListItems) == 0 {
				return fmt.Errorf("element %d: list requires list_items", i)
			}
		case "link":
			if elem.URL == "" || elem.Text == "" {
				return fmt.Errorf("element %d: link requires url and text", i)
			}
		}
	}

	return nil
}
