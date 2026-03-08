package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

// DocumentTemplate represents the root structure of a PDF template
type DocumentTemplate struct {
	// "canvas" (default) or "flow"
	Mode string `json:"mode,omitempty"`

	// Page settings
	PageSize    string  `json:"page_size,omitempty"`   // A4, A3, Letter, Legal, or custom
	PageWidth   float64 `json:"page_width,omitempty"`  // Custom width in points (if PageSize not set)
	PageHeight  float64 `json:"page_height,omitempty"` // Custom height in points
	Orientation string  `json:"orientation,omitempty"` // portrait or landscape
	Margin      *Margin `json:"margin,omitempty"`
	Background  *Color  `json:"background,omitempty"` // Global page background color

	// Grid system for precise layout (Canvas Mode only)
	Grid *Grid `json:"grid,omitempty"`

	// Global Watermark
	Watermark *Watermark `json:"watermark,omitempty"`

	// Document metadata
	Title   string `json:"title,omitempty"`
	Author  string `json:"author,omitempty"`
	Subject string `json:"subject,omitempty"`
	Creator string `json:"creator,omitempty"`

	// Default font settings
	DefaultFont *FontConfig `json:"default_font,omitempty"`

	// Font registrations
	Fonts []FontDef `json:"fonts,omitempty"`

	// ---------------------------------------------------------
	// CANVAS MODE ELEMENTS (Fixed Layouts)
	// ---------------------------------------------------------
	Header   []Element `json:"header,omitempty"`
	Footer   []Element `json:"footer,omitempty"`
	Elements []Element `json:"elements,omitempty"`

	// ---------------------------------------------------------
	// FLOW MODE ELEMENTS (Dynamic Layouts via Maroto)
	// ---------------------------------------------------------
	FlowContent []FlowRow `json:"flow_content,omitempty"`
}

// FlowRow represents a dynamic row in Flow mode
type FlowRow struct {
	Height  float64      `json:"height,omitempty"` // Optional fixed height, 0 = auto
	Columns []FlowColumn `json:"columns"`
}

// FlowColumn represents a column within a row (1-12 grid system)
type FlowColumn struct {
	Size  int        `json:"size"` // 1 to 12
	Style *FlowStyle `json:"style,omitempty"`
	Text  string     `json:"text,omitempty"`
	Image *FlowImage `json:"image,omitempty"`
	// Support for embedded components
	Components []FlowComponent `json:"components,omitempty"`
}

// FlowComponent represents a component within a flow column
type FlowComponent struct {
	Type       string      `json:"type"` // text, image, signature, barcode, qrcode
	Text       string      `json:"text,omitempty"`
	Image      *FlowImage  `json:"image,omitempty"`
	Style      *FlowStyle  `json:"style,omitempty"`
	Properties interface{} `json:"properties,omitempty"`
}

type FlowStyle struct {
	Font         *FontConfig `json:"font,omitempty"`
	Alignment    string      `json:"alignment,omitempty"` // left, center, right, justify
	Vertical     string      `json:"vertical,omitempty"`  // top, middle, bottom
	TextColor    *Color      `json:"text_color,omitempty"`
	BgColor      *Color      `json:"background_color,omitempty"`
	BorderColor  *Color      `json:"border_color,omitempty"`
	BorderWidth  float64     `json:"border_width,omitempty"`
	BorderType   string      `json:"border_type,omitempty"` // solid, dashed, dotted
	Padding      *Padding    `json:"padding,omitempty"`
	BorderRadius float64     `json:"border_radius,omitempty"` // For card-like styling
}

type FlowImage struct {
	Path      string  `json:"path,omitempty"`
	URL       string  `json:"url,omitempty"`
	Base64    string  `json:"base64,omitempty"`
	Width     float64 `json:"width,omitempty"`
	Height    float64 `json:"height,omitempty"`
	Extension string  `json:"extension,omitempty"` // png, jpg, jpeg
}

// --- Common Structures ---

type Grid struct {
	Size float64 `json:"size"`
	Draw bool    `json:"draw,omitempty"`
}

type Watermark struct {
	Text    string      `json:"text"`
	Font    *FontConfig `json:"font,omitempty"`
	Opacity float64     `json:"opacity,omitempty"`
}

type Margin struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

type FontConfig struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Style  string  `json:"style,omitempty"`
	Color  *Color  `json:"color,omitempty"`
}

type FontDef struct {
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	Data     []byte `json:"-"`
}

type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a,omitempty"` // Alpha for transparency support
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height,omitempty"`
}

type Border struct {
	Top    bool    `json:"top,omitempty"`
	Bottom bool    `json:"bottom,omitempty"`
	Left   bool    `json:"left,omitempty"`
	Right  bool    `json:"right,omitempty"`
	All    bool    `json:"all,omitempty"`
	Style  string  `json:"style,omitempty"` // solid, dashed, dotted
	Width  float64 `json:"width,omitempty"` // Border width
}

type Alignment struct {
	Horizontal string `json:"horizontal,omitempty"`
	Vertical   string `json:"vertical,omitempty"`
}

// Element represents a PDF element
type Element struct {
	Type            string        `json:"type"`
	Position        *Position     `json:"position,omitempty"`
	Size            *Size         `json:"size,omitempty"`
	ZIndex          int           `json:"z_index,omitempty"`
	Opacity         float64       `json:"opacity,omitempty"`
	Text            string        `json:"text,omitempty"`
	Font            *FontConfig   `json:"font,omitempty"`
	Alignment       *Alignment    `json:"alignment,omitempty"`
	LineHeight      float64       `json:"line_height,omitempty"`
	RTL             bool          `json:"rtl,omitempty"`
	ImagePath       string        `json:"image_path,omitempty"`
	ImageData       []byte        `json:"-"`
	ImageURL        string        `json:"image_url,omitempty"`
	FillColor       *Color        `json:"fill_color,omitempty"`
	LineColor       *Color        `json:"line_color,omitempty"`
	LineWidth       float64       `json:"line_width,omitempty"`
	LineStyle       string        `json:"line_style,omitempty"`
	Columns         []TableColumn `json:"columns,omitempty"`
	Rows            []TableRow    `json:"rows,omitempty"`
	Header          *TableSection `json:"header,omitempty"`
	CellPadding     *Padding      `json:"cell_padding,omitempty"`
	Border          *Border       `json:"border,omitempty"`
	BorderColor     *Color        `json:"border_color,omitempty"`
	BackgroundColor *Color        `json:"background_color,omitempty"`
	EndX            float64       `json:"end_x,omitempty"`
	EndY            float64       `json:"end_y,omitempty"`
	ListItems       []string      `json:"list_items,omitempty"`
	ListType        string        `json:"list_type,omitempty"`
	URL             string        `json:"url,omitempty"`
	Height          float64       `json:"height,omitempty"`
	Indent          float64       `json:"indent,omitempty"`

	// ---------------------------------------------------------
	// Card Element Support
	// ---------------------------------------------------------
	Elements []Element `json:"elements,omitempty"` // Child elements for card/container
	Padding  *Padding  `json:"padding,omitempty"`  // Internal padding for card
	Shadow   *Shadow   `json:"shadow,omitempty"`   // Shadow effect for card
	Radius   float64   `json:"radius,omitempty"`   // Border radius for card
	Overflow string    `json:"overflow,omitempty"` // hidden, visible (for clipping)

	// ---------------------------------------------------------
	// Enhanced Table Features
	// ---------------------------------------------------------
	Caption      string        `json:"caption,omitempty"`        // Table caption
	CaptionStyle *FontConfig   `json:"caption_style,omitempty"`  // Caption font styling
	Summary      string        `json:"summary,omitempty"`        // Accessibility summary
	TableLayout  string        `json:"table_layout,omitempty"`   // fixed, auto
	Width        float64       `json:"width,omitempty"`          // Table width (0 = auto)
	MinRowHeight float64       `json:"min_row_height,omitempty"` // Minimum row height
	RowHeights   []float64     `json:"row_heights,omitempty"`    // Per-row heights
	Footer       *TableSection `json:"footer,omitempty"`         // Table footer (tfoot)

	// ---------------------------------------------------------
	// Advanced Shape Features
	// ---------------------------------------------------------
	CornerRadius float64 `json:"corner_radius,omitempty"` // Rounded corners for rectangles
	StartAngle   float64 `json:"start_angle,omitempty"`   // For arc/pie
	EndAngle     float64 `json:"end_angle,omitempty"`     // For arc/pie
	Rotation     float64 `json:"rotation,omitempty"`      // Rotation angle in degrees
}

// Shadow defines shadow properties for cards and elements
type Shadow struct {
	Color   *Color  `json:"color,omitempty"`
	OffsetX float64 `json:"offset_x,omitempty"`
	OffsetY float64 `json:"offset_y,omitempty"`
	Blur    float64 `json:"blur,omitempty"`
	Spread  float64 `json:"spread,omitempty"`
	Inset   bool    `json:"inset,omitempty"`
}

type TableColumn struct {
	Width    float64 `json:"width"`
	Align    string  `json:"align,omitempty"`
	MinWidth float64 `json:"min_width,omitempty"`
	MaxWidth float64 `json:"max_width,omitempty"`
}

type TableRow struct {
	Cells       []TableCell `json:"cells"`
	Height      float64     `json:"height,omitempty"`       // Row-specific height
	Background  *Color      `json:"background,omitempty"`   // Row background color
	BorderColor *Color      `json:"border_color,omitempty"` // Row border color
	Border      *Border     `json:"border,omitempty"`       // Row-specific borders
	MinHeight   float64     `json:"min_height,omitempty"`   // Minimum row height
	Repeat      bool        `json:"repeat,omitempty"`       // Repeat row on new pages
}

// TableCell with HTML-table-like control
type TableCell struct {
	Text          string      `json:"text"`
	Font          *FontConfig `json:"font,omitempty"`
	Align         string      `json:"align,omitempty"`
	VerticalAlign string      `json:"vertical_align,omitempty"` // top, middle, bottom
	ColSpan       int         `json:"col_span,omitempty"`
	RowSpan       int         `json:"row_span,omitempty"`
	Background    *Color      `json:"background,omitempty"`
	RTL           bool        `json:"rtl,omitempty"`
	Border        *Border     `json:"border,omitempty"`       // Cell-specific borders
	BorderColor   *Color      `json:"border_color,omitempty"` // Cell border color
	Padding       *Padding    `json:"padding,omitempty"`      // Cell-specific padding
	Width         float64     `json:"width,omitempty"`        // Cell width override
	Height        float64     `json:"height,omitempty"`       // Cell height
	TextWrap      bool        `json:"text_wrap,omitempty"`    // Text wrapping (default true)
	Overflow      string      `json:"overflow,omitempty"`     // hidden, visible, ellipsis
	// Support nested content
	Elements []Element `json:"elements,omitempty"` // Nested elements in cell
}

// TableSection for table header (thead) or footer (tfoot) support
type TableSection struct {
	Cells       []TableCell `json:"cells"`
	Font        *FontConfig `json:"font,omitempty"`
	Background  *Color      `json:"background,omitempty"`
	Border      *Border     `json:"border,omitempty"`
	BorderColor *Color      `json:"border_color,omitempty"`
	Height      float64     `json:"height,omitempty"`
	Repeat      bool        `json:"repeat,omitempty"` // Repeat header/footer on each page
}

type TableHeader = TableSection
type TableFooter = TableSection

type Padding struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// ParseTemplateBytes parses a JSON template from bytes
func ParseTemplateBytes(data []byte) (*DocumentTemplate, error) {
	var tmpl DocumentTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parsing template JSON: %w", err)
	}
	applyTemplateDefaults(&tmpl)
	return &tmpl, nil
}

// ParseTemplate parses from reader
func ParseTemplate(r io.Reader) (*DocumentTemplate, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}
	return ParseTemplateBytes(data)
}

func applyTemplateDefaults(tmpl *DocumentTemplate) {
	if tmpl.Mode == "" {
		tmpl.Mode = "canvas"
	}
	if tmpl.PageSize == "" && tmpl.PageWidth == 0 {
		tmpl.PageSize = "A4"
	}
	if tmpl.Orientation == "" {
		tmpl.Orientation = "portrait"
	}
	if tmpl.Margin == nil {
		tmpl.Margin = &Margin{Top: 50, Bottom: 50, Left: 50, Right: 50}
	} else {
		if tmpl.Margin.Top < 0 {
			tmpl.Margin.Top = 0
		}
		if tmpl.Margin.Bottom < 0 {
			tmpl.Margin.Bottom = 0
		}
		if tmpl.Margin.Left < 0 {
			tmpl.Margin.Left = 0
		}
		if tmpl.Margin.Right < 0 {
			tmpl.Margin.Right = 0
		}
	}
	if tmpl.DefaultFont == nil {
		tmpl.DefaultFont = &FontConfig{Family: "Helvetica", Size: 12}
	}
}

// Validate validates the template
func (t *DocumentTemplate) Validate() error {
	if t.Mode == "flow" {
		if len(t.FlowContent) == 0 {
			return fmt.Errorf("flow mode requires flow_content")
		}
		return nil
	}
	if len(t.Elements) == 0 && len(t.Header) == 0 && len(t.Footer) == 0 {
		return fmt.Errorf("canvas mode requires elements, header, or footer")
	}
	return nil
}
