package generator

import (
	"os"
	"testing"

	"github.com/amsaid/gopdf-generator/pkg/parser"
)

func TestNew(t *testing.T) {
	// Test with default config
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}
	defer gen.Close()

	if gen.pdf == nil {
		t.Error("Expected pdf to be initialized")
	}
	if gen.fontMgr == nil {
		t.Error("Expected fontMgr to be initialized")
	}

	// Test with custom config
	config := &Config{
		FontDir:    "./test-fonts",
		TempDir:    os.TempDir(),
		EmbedFonts: true,
	}

	gen2, err := New(config)
	if err != nil {
		t.Fatalf("New(config) error = %v", err)
	}
	defer gen2.Close()

	// Cleanup test directory
	os.RemoveAll("./test-fonts")
}

func TestPDFGenerator_Generate(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	tests := []struct {
		name    string
		tmpl    *parser.DocumentTemplate
		wantErr bool
	}{
		{
			name: "simple text document",
			tmpl: &parser.DocumentTemplate{
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{Type: "text", Text: "Hello World"},
				},
			},
			wantErr: false,
		},
		{
			name: "document with table",
			tmpl: &parser.DocumentTemplate{
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{
						Type: "table",
						Columns: []parser.TableColumn{
							{Width: 100},
							{Width: 100},
						},
						Rows: []parser.TableRow{
							{
								Cells: []parser.TableCell{
									{Text: "Cell 1"},
									{Text: "Cell 2"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "document with shapes",
			tmpl: &parser.DocumentTemplate{
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{
						Type: "rect",
						Size: &parser.Size{Width: 100, Height: 50},
						FillColor: &parser.Color{R: 255, G: 0, B: 0},
					},
					{
						Type: "ellipse",
						Size: &parser.Size{Width: 80, Height: 60},
					},
					{
						Type: "line",
						EndX: 200,
						EndY: 100,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "document with styled text",
			tmpl: &parser.DocumentTemplate{
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{
						Type: "text",
						Text: "Colored Text",
						Font: &parser.FontConfig{
							Family: "Helvetica",
							Size:   18,
							Style:  "B",
							Color:  &parser.Color{R: 41, G: 128, B: 185},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty elements",
			tmpl: &parser.DocumentTemplate{
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				Elements: []parser.Element{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := gen.Generate(tt.tmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify PDF was generated (should start with %PDF)
			if buf.Len() == 0 {
				t.Error("Generated PDF is empty")
			}

			data := buf.Bytes()
			if len(data) < 4 || string(data[:4]) != "%PDF" {
				t.Error("Generated data doesn't appear to be a valid PDF")
			}
		})
	}
}

func TestPDFGenerator_GenerateFromJSON(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	json := `{
		"page_size": "A4",
		"margin": {"top": 50, "bottom": 50, "left": 50, "right": 50},
		"default_font": {"family": "Helvetica", "size": 12},
		"elements": [
			{"type": "text", "text": "JSON Test"}
		]
	}`

	buf, err := gen.GenerateFromJSON([]byte(json))
	if err != nil {
		t.Fatalf("GenerateFromJSON() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Generated PDF is empty")
	}
}

func TestPDFGenerator_GenerateToFile(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	tmpl := &parser.DocumentTemplate{
		PageSize: "A4",
		Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
		DefaultFont: &parser.FontConfig{
			Family: "Helvetica",
			Size:   12,
		},
		Elements: []parser.Element{
			{Type: "text", Text: "File Test"},
		},
	}

	outputPath := "./test-output/test.pdf"
	os.MkdirAll("./test-output", 0755)
	defer os.RemoveAll("./test-output")

	err = gen.GenerateToFile(tmpl, outputPath)
	if err != nil {
		t.Fatalf("GenerateToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify file is a valid PDF
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Reading output file: %v", err)
	}

	if len(data) < 4 || string(data[:4]) != "%PDF" {
		t.Error("Output file is not a valid PDF")
	}
}

func TestPDFGenerator_PageSizes(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	pageSizes := []string{"A4", "A3", "A5", "Letter", "Legal"}

	for _, size := range pageSizes {
		t.Run(size, func(t *testing.T) {
			tmpl := &parser.DocumentTemplate{
				PageSize: size,
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{Type: "text", Text: "Page Size Test"},
				},
			}

			buf, err := gen.Generate(tmpl)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			if buf.Len() == 0 {
				t.Error("Generated PDF is empty")
			}
		})
	}
}

func TestPDFGenerator_Orientations(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	orientations := []string{"portrait", "landscape"}

	for _, orientation := range orientations {
		t.Run(orientation, func(t *testing.T) {
			tmpl := &parser.DocumentTemplate{
				PageSize:    "A4",
				Orientation: orientation,
				Margin:      &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				DefaultFont: &parser.FontConfig{
					Family: "Helvetica",
					Size:   12,
				},
				Elements: []parser.Element{
					{Type: "text", Text: "Orientation Test"},
				},
			}

			buf, err := gen.Generate(tmpl)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			if buf.Len() == 0 {
				t.Error("Generated PDF is empty")
			}
		})
	}
}

func TestPDFGenerator_Metadata(t *testing.T) {
	gen, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	tmpl := &parser.DocumentTemplate{
		Title:       "Test Title",
		Author:      "Test Author",
		Subject:     "Test Subject",
		Creator:     "Test Creator",
		PageSize:    "A4",
		Margin:      &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
		DefaultFont: &parser.FontConfig{Family: "Helvetica", Size: 12},
		Elements:    []parser.Element{{Type: "text", Text: "Test"}},
	}

	buf, err := gen.Generate(tmpl)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Generated PDF is empty")
	}

	// Note: We can't easily verify metadata without parsing the PDF,
	// but we can verify the PDF was generated successfully
}