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

	// Example 2: Generate PDF programmatically
	fmt.Println("\n=== Example 2: Generate programmatically ===")
	generateProgrammatically()

	// Example 3: Generate with custom fonts
	fmt.Println("\n=== Example 3: Generate with custom fonts ===")
	generateWithCustomFonts()

	// Example 4: Generate with RTL text
	fmt.Println("\n=== Example 4: Generate with RTL text ===")
	generateWithRTL()
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
			},
			{
				Type:     "cell",
				Position: &parser.Position{X: 60, Y: 300},
				Size:     &parser.Size{Width: 180, Height: 50},
				Text:     "Success!",
				Font: &parser.FontConfig{
					Size:  16,
					Style: "B",
					Color: &parser.Color{R: 255, G: 255, B: 255},
				},
				Alignment:       &parser.Alignment{Horizontal: "C", Vertical: "M"},
				BackgroundColor: &parser.Color{R: 46, G: 204, B: 113},
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

	// Register a custom font (if available)
	// For this example, we'll use standard fonts
	// In practice, you would have .ttf files in the fonts directory

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
