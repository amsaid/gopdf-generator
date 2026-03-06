package main

import (
	"fmt"
	"log"
	"os"

	"github.com/amsaid/gopdf-generator/pkg/generator"
	"github.com/amsaid/gopdf-generator/pkg/parser"
)

func main() {
	// Example 1: Generate PDF from JSON template file
	fmt.Println("=== Example 1: Generate from JSON file ===")
	generateFromJSONFile()
	generateFromComplexJSONFile()

	// Example 2: Generate PDF programmatically
	fmt.Println("\n=== Example 2: Generate programmatically ===")
	generateProgrammatically()
	generateGridDashboard()

	// Example 3: Generate with custom fonts
	fmt.Println("\n=== Example 3: Generate with custom fonts ===")
	generateWithCustomFonts()

	// Example 4: Generate with RTL text
	fmt.Println("\n=== Example 4: Generate with RTL text ===")
	generateWithRTL()

	// Example 4: Generate with RTL text
	fmt.Println("\n=== Example 5: More json exemples ===")
	generateMoreExemples()

}

// generateFromJSONFile generates PDF from a JSON template file
func generateFromJSONFile() {
	// Create generator
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}

	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}

	fontsList := gen.GetFontManager().ListFonts()
	log.Printf("Available fonts: %v", fontsList)
	defer gen.Close()

	// Read template file
	templateData, err := os.ReadFile("./examples/basic_invoice.json")
	if err != nil {
		log.Printf("Error reading template: %v", err)
		return
	}

	// Generate PDF
	buf, err := gen.GenerateFromJSON(templateData)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("output_invoice.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: output_invoice.pdf")
}

func generateFromComplexJSONFile() {
	// Create generator
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}

	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}

	fontsList := gen.GetFontManager().ListFonts()
	log.Printf("Available fonts: %v", fontsList)
	defer gen.Close()

	// Read template file
	templateData, err := os.ReadFile("./examples/complex_report.json")
	if err != nil {
		log.Printf("Error reading template: %v", err)
		return
	}

	// Generate PDF
	buf, err := gen.GenerateFromJSON(templateData)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("output_complex_report.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: output_invoice.pdf")
}

func generateMoreExemples() {
	autoGen("dashboard")
	autoGen("certificate")
	autoGen("idcard")
	autoGen("flow_invoice")
	autoGen("flow_report_rtl")
	autoGen("fow_article")
	autoGen("emploi")
}
func autoGen(template string) {
	// Create generator
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}

	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}
	fontsList := gen.GetFontManager().ListFonts()
	log.Printf("Available fonts: %v", fontsList)
	defer gen.Close()

	// Read template file
	templateData, err := os.ReadFile(fmt.Sprintf("./examples/%s.json", template))
	if err != nil {
		log.Printf("Error reading template: %v", err)
		return
	}

	// Generate PDF
	buf, err := gen.GenerateFromJSON(templateData)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile(fmt.Sprintf("%s.pdf", template), buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println(fmt.Sprintf("✓ Generated: %s.pdf", template))
}

// generateProgrammatically creates a PDF programmatically
func generateProgrammatically() {
	// Create generator
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}
	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}
	defer gen.Close()

	// Build template programmatically
	template := &parser.DocumentTemplate{
		Title:       "Programmatic PDF",
		Author:      "GoPDF Generator",
		PageSize:    "A4",
		Orientation: "portrait",
		Margin: &parser.Margin{
			Top:    50,
			Bottom: 50,
			Left:   50,
			Right:  50,
		},
		DefaultFont: &parser.FontConfig{
			Family: "Helvetica",
			Size:   12,
		},
		Elements: []parser.Element{
			{
				Type: "text",
				Text: "Hello, World!",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   24,
					Style:  "B",
					Color:  &parser.Color{R: 41, G: 128, B: 185},
				},
				Alignment: &parser.Alignment{Horizontal: "C"},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type:       "text",
				Text:       "This PDF was generated programmatically using GoPDF Generator library.",
				LineHeight: 18,
			},
			{
				Type:   "newline",
				Height: 30,
			},
			{
				Type:     "list",
				ListType: "ul",
				ListItems: []string{
					"Supports Bulleted and Numbered Lists",
					"Clickable Hyperlinks",
					"Base64 Embedded Images",
				},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "link",
				Text: "Check out our website",
				URL:  "https://github.com/signintech/gopdf",
			},
			{
				Type:   "newline",
				Height: 30,
			},
			{
				Type: "table",
				Columns: []parser.TableColumn{
					{Width: 200},
					{Width: 100},
					{Width: 100},
				},
				Header: &parser.TableHeader{
					Cells: []parser.TableCell{
						{Text: "Item", Font: &parser.FontConfig{Style: "B"}},
						{Text: "Quantity", Align: "C", Font: &parser.FontConfig{Style: "B"}},
						{Text: "Price", Align: "R", Font: &parser.FontConfig{Style: "B"}},
					},
					Background: &parser.Color{R: 52, G: 73, B: 94},
				},
				Rows: []parser.TableRow{
					{
						Cells: []parser.TableCell{
							{Text: "Product A"},
							{Text: "5", Align: "C"},
							{Text: "$10.00", Align: "R"},
						},
					},
					{
						Cells: []parser.TableCell{
							{Text: "Product B"},
							{Text: "3", Align: "C"},
							{Text: "$25.00", Align: "R"},
						},
					},
					{
						Cells: []parser.TableCell{
							{Text: "Product C"},
							{Text: "1", Align: "C"},
							{Text: "$100.00", Align: "R"},
						},
					},
				},
				CellPadding: &parser.Padding{Top: 8, Bottom: 8, Left: 10, Right: 10},
				Border:      &parser.Border{All: true},
			},
			{
				Type:   "newline",
				Height: 30,
			},
			{
				Type:      "rect",
				Size:      &parser.Size{Width: 200, Height: 60},
				FillColor: &parser.Color{R: 46, G: 204, B: 113},
				LineColor: &parser.Color{R: 39, G: 174, B: 96},
				LineWidth: 2,
				ZIndex:    1,
			},
			{
				Type:     "cell",
				Position: &parser.Position{X: 60, Y: 420},
				Size:     &parser.Size{Width: 180, Height: 50},
				Text:     "Success!",
				Font: &parser.FontConfig{
					Size:  16,
					Style: "B",
					Color: &parser.Color{R: 255, G: 255, B: 255},
				},
				Alignment:       &parser.Alignment{Horizontal: "C", Vertical: "M"},
				BackgroundColor: &parser.Color{R: 46, G: 204, B: 113},
				Opacity:         0.85,
				ZIndex:          2,
			},
		},
	}

	// Generate PDF
	buf, err := gen.Generate(template)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("output_programmatic.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: output_programmatic.pdf")

}

func generateGridDashboard() {
	// Create generator
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}
	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}
	defer gen.Close()

	// Define brand colors
	bgGray := &parser.Color{R: 240, G: 242, B: 245}
	cardWhite := &parser.Color{R: 255, G: 255, B: 255}
	textDark := &parser.Color{R: 40, G: 40, B: 40}
	textWhite := &parser.Color{R: 255, G: 255, B: 255}
	primaryBlue := &parser.Color{R: 52, G: 152, B: 219}
	successGreen := &parser.Color{R: 46, G: 204, B: 113}
	warningOrange := &parser.Color{R: 243, G: 156, B: 18}

	template := &parser.DocumentTemplate{
		Title:       "Smart Grid Dashboard",
		PageSize:    "A4",
		Orientation: "landscape",
		Background:  bgGray,
		Grid: &parser.Grid{
			Size: 20,   // Every cell is 20x20 points
			Draw: true, // Keep TRUE to see the grid lines perfectly encasing our sloppy inputs
		},
		Margin:      &parser.Margin{Top: 20, Bottom: 20, Left: 20, Right: 20},
		DefaultFont: &parser.FontConfig{Family: "Helvetica", Size: 12, Color: textDark},
		Elements: []parser.Element{

			// ==========================================
			// HEADER AREA
			// ==========================================
			{
				Type: "text",
				Text: "Executive Performance Dashboard (Smart Grid)",
				Font: &parser.FontConfig{Family: "Helvetica", Size: 24, Style: "B", Color: textDark},
				// SHOWCASE 1: Sloppy Position. 38 -> snaps to 40, 43 -> snaps to 40.
				Position: &parser.Position{X: 38, Y: 43},
			},

			// ==========================================
			// ROW 1: KPI CARDS (Target Y = 80, W = 220, H = 80)
			// ==========================================

			// KPI Card 1: Revenue (Blue)
			{
				Type: "rect",
				// SHOWCASE 2: Sloppy Sizes. 217 -> snaps to 220, 84 -> snaps to 80.
				Size:      &parser.Size{Width: 217, Height: 84},
				Position:  &parser.Position{X: 41, Y: 78}, // Snaps to 40, 80
				FillColor: primaryBlue,
				ZIndex:    1,
			},
			{
				Type: "text", Text: "Total Revenue",
				Font:     &parser.FontConfig{Size: 12, Color: textWhite},
				Position: &parser.Position{X: 58, Y: 102}, // Snaps to 60, 100
				ZIndex:   2,
			},
			{
				Type: "text", Text: "$124,500",
				Font:     &parser.FontConfig{Size: 24, Style: "B", Color: textWhite},
				Position: &parser.Position{X: 61, Y: 118}, // Snaps to 60, 120
				ZIndex:   2,
			},

			// KPI Card 2: Users (Green)
			{
				Type:      "rect",
				Size:      &parser.Size{Width: 223, Height: 77}, // Snaps to 220, 80
				Position:  &parser.Position{X: 282, Y: 81},      // Snaps to 280, 80
				FillColor: successGreen,
				ZIndex:    1,
			},
			{
				Type: "text", Text: "Active Users",
				Font:     &parser.FontConfig{Size: 12, Color: textWhite},
				Position: &parser.Position{X: 303, Y: 98}, // Snaps to 300, 100
				ZIndex:   2,
			},
			{
				Type: "text", Text: "8,432",
				Font:     &parser.FontConfig{Size: 24, Style: "B", Color: textWhite},
				Position: &parser.Position{X: 301, Y: 122}, // Snaps to 300, 120
				ZIndex:   2,
			},

			// KPI Card 3: Conversion (Orange)
			{
				Type:      "rect",
				Size:      &parser.Size{Width: 216, Height: 83}, // Snaps to 220, 80
				Position:  &parser.Position{X: 518, Y: 79},      // Snaps to 520, 80
				FillColor: warningOrange,
				ZIndex:    1,
			},
			{
				Type: "text", Text: "Conversion Rate",
				Font:     &parser.FontConfig{Size: 12, Color: textWhite},
				Position: &parser.Position{X: 538, Y: 101}, // Snaps to 540, 100
				ZIndex:   2,
			},
			{
				Type: "text", Text: "4.2%",
				Font:     &parser.FontConfig{Size: 24, Style: "B", Color: textWhite},
				Position: &parser.Position{X: 541, Y: 119}, // Snaps to 540, 120
				ZIndex:   2,
			},

			// ==========================================
			// ROW 2: CHARTS & TABLES (Target Y = 180)
			// ==========================================

			// Panel 1: Bar Chart Background Card
			{
				Type:      "rect",
				Size:      &parser.Size{Width: 462, Height: 258}, // Snaps to 460, 260
				Position:  &parser.Position{X: 39, Y: 182},       // Snaps to 40, 180
				FillColor: cardWhite,
				ZIndex:    1,
			},
			{
				Type: "text", Text: "Monthly Sales Growth",
				Font:     &parser.FontConfig{Size: 14, Style: "B"},
				Position: &parser.Position{X: 58, Y: 202}, // Snaps to 60, 200
				ZIndex:   2,
			},
			// "Simulated" Bar Chart. Width of 38 snaps perfectly to 40.
			{Type: "rect", Size: &parser.Size{Width: 38, Height: 79}, Position: &parser.Position{X: 78, Y: 341}, FillColor: primaryBlue, ZIndex: 2},
			{Type: "rect", Size: &parser.Size{Width: 42, Height: 121}, Position: &parser.Position{X: 141, Y: 298}, FillColor: primaryBlue, ZIndex: 2},
			{Type: "rect", Size: &parser.Size{Width: 39, Height: 158}, Position: &parser.Position{X: 198, Y: 262}, FillColor: primaryBlue, ZIndex: 2},
			{Type: "rect", Size: &parser.Size{Width: 41, Height: 102}, Position: &parser.Position{X: 262, Y: 318}, FillColor: primaryBlue, ZIndex: 2},
			{Type: "rect", Size: &parser.Size{Width: 37, Height: 181}, Position: &parser.Position{X: 319, Y: 239}, FillColor: primaryBlue, ZIndex: 2},
			{Type: "rect", Size: &parser.Size{Width: 43, Height: 202}, Position: &parser.Position{X: 381, Y: 218}, FillColor: primaryBlue, ZIndex: 2},

			// Panel 2: Table Background Card
			{
				Type:      "rect",
				Size:      &parser.Size{Width: 261, Height: 259}, // Snaps to 260, 260
				Position:  &parser.Position{X: 518, Y: 181},      // Snaps to 520, 180
				FillColor: cardWhite,
				ZIndex:    1,
			},
			{
				Type: "text", Text: "Recent Transactions",
				Font:     &parser.FontConfig{Size: 14, Style: "B"},
				Position: &parser.Position{X: 541, Y: 199}, // Snaps to 540, 200
				ZIndex:   2,
			},
			// An absolutely positioned Table snapped to the grid inside the card
			{
				Type:     "table",
				Position: &parser.Position{X: 538, Y: 242}, // Snaps to 540, 240
				// SHOWCASE 3: Table width constraint snaps to 220 so it fits the card beautifully
				Size:   &parser.Size{Width: 224},
				ZIndex: 2,
				Columns: []parser.TableColumn{
					{Width: 120, Align: "L"},
					{Width: 100, Align: "R"},
				},
				Header: &parser.TableHeader{
					Background: &parser.Color{R: 230, G: 230, B: 230},
					Cells: []parser.TableCell{
						{Text: "Client"},
						{Text: "Amount"},
					},
				},
				Rows: []parser.TableRow{
					{Cells: []parser.TableCell{{Text: "Acme Corp"}, {Text: "$1,200"}}},
					{Cells: []parser.TableCell{{Text: "Globex"}, {Text: "$850"}}},
					{Cells: []parser.TableCell{{Text: "Initech"}, {Text: "$3,400"}}},
					{Cells: []parser.TableCell{{Text: "Soylent"}, {Text: "$150"}}},
					{Cells: []parser.TableCell{{Text: "Umbrella"}, {Text: "$9,990"}}},
				},
			},
		},
	}

	// Generate PDF
	buf, err := gen.Generate(template)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("dashboard_smart_grid.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: dashboard_smart_grid.pdf")
}

// generateWithCustomFonts demonstrates using custom fonts
func generateWithCustomFonts() {
	// Create generator with custom config
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}

	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}

	fontsList := gen.GetFontManager().ListFonts()
	log.Printf("Available fonts: %v", fontsList)
	defer gen.Close()

	template := &parser.DocumentTemplate{
		Title:       "Custom Font Example",
		PageSize:    "A4",
		Orientation: "portrait",
		Margin: &parser.Margin{
			Top: 50, Bottom: 50, Left: 50, Right: 50,
		},
		DefaultFont: &parser.FontConfig{
			Family: "Helvetica",
			Size:   12,
		},
		Elements: []parser.Element{
			{
				Type: "text",
				Text: "Standard Fonts Demo",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   20,
					Style:  "B",
				},
				Alignment: &parser.Alignment{Horizontal: "C"},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "text",
				Text: "Helvetica (Normal)",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   14,
				},
			},
			{
				Type: "text",
				Text: "Helvetica Bold",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   14,
					Style:  "B",
				},
			},
			{
				Type: "text",
				Text: "Helvetica Oblique",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   14,
					Style:  "I",
				},
			},
			{
				Type: "text",
				Text: "Helvetica Bold Oblique",
				Font: &parser.FontConfig{
					Family: "Helvetica",
					Size:   14,
					Style:  "BI",
				},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "text",
				Text: "Times Roman (Serif Font)",
				Font: &parser.FontConfig{
					Family: "Helvetica", // Changed from Times-Roman
					Size:   14,
				},
			},
			{
				Type: "text",
				Text: "Times Bold Italic",
				Font: &parser.FontConfig{
					Family: "Helvetica", // Changed from Times-Roman
					Size:   14,
					Style:  "BI",
				},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "text",
				Text: "Courier (Monospace Font)",
				Font: &parser.FontConfig{
					Family: "Courier",
					Size:   14,
				},
			},
			{
				Type: "text",
				Text: "Courier Bold",
				Font: &parser.FontConfig{
					Family: "Courier",
					Size:   14,
					Style:  "B",
				},
			},
		},
	}

	// Generate PDF
	buf, err := gen.Generate(template)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("output_fonts.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: output_fonts.pdf")
}

// generateWithRTL demonstrates RTL text support
func generateWithRTL() {
	config := &generator.Config{
		FontDir: "./fonts",
		TempDir: os.TempDir(),
	}
	gen, err := generator.New(config)
	if err != nil {
		log.Printf("Error creating generator: %v", err)
		return
	}

	defer gen.Close()

	template := &parser.DocumentTemplate{
		Title:       "RTL Example",
		PageSize:    "A4",
		Orientation: "portrait",
		Margin: &parser.Margin{
			Top: 50, Bottom: 50, Left: 50, Right: 50,
		},
		DefaultFont: &parser.FontConfig{
			Family: "NotoSansArabic",
			Size:   12,
		},
		Elements: []parser.Element{
			{
				Type: "text",
				Text: "RTL (Right-to-Left) Text Support",
				Font: &parser.FontConfig{
					Size:  18,
					Style: "B",
					Color: &parser.Color{R: 41, G: 128, B: 185},
				},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type:       "text",
				Text:       "The RTL flag tells the generator to process text for right-to-left languages like Arabic, Hebrew, Persian, and Urdu.",
				LineHeight: 18,
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "cell",
				Size: &parser.Size{Width: 400, Height: 50},
				Text: "Sample RTL Text Cell (requires Unicode font)",
				RTL:  true,
				Alignment: &parser.Alignment{
					Horizontal: "R",
				},
				BackgroundColor: &parser.Color{R: 245, G: 245, B: 245},
				Border:          &parser.Border{All: true},
			},
			{
				Type:   "newline",
				Height: 20,
			},
			{
				Type: "table",
				Columns: []parser.TableColumn{
					{Width: 200},
					{Width: 200},
				},
				Header: &parser.TableHeader{
					Cells: []parser.TableCell{
						{Text: "English", Font: &parser.FontConfig{Style: "B"}},
						{Text: "RTL Language", Font: &parser.FontConfig{Style: "B"}, RTL: true},
					},
					Background: &parser.Color{R: 52, G: 73, B: 94},
				},
				Rows: []parser.TableRow{
					{
						Cells: []parser.TableCell{
							{Text: "Hello World"},
							{Text: "مرحبا بالعالم", RTL: true, Font: &parser.FontConfig{Family: "Arial"}},
						},
					},
					{
						Cells: []parser.TableCell{
							{Text: "Thank you"},
							{Text: "شكرا لك", RTL: true},
						},
					},
				},
				CellPadding: &parser.Padding{Top: 10, Bottom: 10, Left: 10, Right: 10},
				Border:      &parser.Border{All: true},
			},
		},
	}

	// Generate PDF
	buf, err := gen.Generate(template)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile("output_rtl.pdf", buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving PDF: %v", err)
		return
	}

	fmt.Println("✓ Generated: output_rtl.pdf")
	fmt.Println("  Note: For proper Arabic/Hebrew text rendering, use a Unicode font like Noto Sans Arabic")
}
