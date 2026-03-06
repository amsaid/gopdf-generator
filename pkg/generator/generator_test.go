package generator

import (
	"testing"

	"github.com/amsaid/gopdf-generator/pkg/parser"
)

func TestNew(t *testing.T) {
	config := DefaultConfig()
	gen, err := New(config)
	if err != nil {
		t.Fatalf("New(config) error = %v", err)
	}
	defer gen.Close()

	if gen.fontMgr == nil {
		t.Error("Expected fontMgr to be initialized")
	}
}

func TestPDFGenerator_Generate(t *testing.T) {
	config := DefaultConfig()
	gen, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	// Register a font that might be used in tests
	_ = gen.RegisterFont("Roboto", "../../fonts/Roboto-Regular.ttf")

	tests := []struct {
		name    string
		tmpl    *parser.DocumentTemplate
		wantErr bool
	}{
		// --- Canvas Mode Tests ---
		{
			name: "canvas: simple text document",
			tmpl: &parser.DocumentTemplate{
				Mode:     "canvas",
				PageSize: "A4",
				Margin:   &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50},
				Elements: []parser.Element{{Type: "text", Text: "Hello World"}},
			},
			wantErr: false,
		},
		{
			name: "canvas: document with zindex and opacity",
			tmpl: &parser.DocumentTemplate{
				Mode:   "canvas",
				Header: []parser.Element{{Type: "text", Text: "Page Header"}},
				Elements: []parser.Element{
					{Type: "rect", Size: &parser.Size{Width: 100, Height: 50}, ZIndex: 2, Opacity: 0.5},
					{Type: "text", Text: "Behind Rect", ZIndex: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "canvas: invalid empty elements",
			tmpl: &parser.DocumentTemplate{
				Mode:     "canvas",
				Elements: []parser.Element{},
			},
			wantErr: true,
		},
		// --- Flow Mode Tests ---
		{
			name: "flow: simple document",
			tmpl: &parser.DocumentTemplate{
				Mode: "flow",
				FlowContent: []parser.FlowRow{
					{Height: 20, Columns: []parser.FlowColumn{{Size: 12, Text: "Hello Flow World"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "flow: document with multiple columns",
			tmpl: &parser.DocumentTemplate{
				Mode: "flow",
				FlowContent: []parser.FlowRow{
					{
						Height: 20,
						Columns: []parser.FlowColumn{
							{Size: 6, Text: "Column 1"},
							{Size: 6, Text: "Column 2"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "flow: document with RTL text",
			tmpl: &parser.DocumentTemplate{
				Mode: "flow",
				FlowContent: []parser.FlowRow{
					{
						Height: 40,
						Columns: []parser.FlowColumn{
							{Size: 12, Text: "مرحبا بالعالم", Style: &parser.FlowStyle{Alignment: "right"}},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply defaults if not set in test
			if tt.tmpl.Margin == nil {
				tt.tmpl.Margin = &parser.Margin{Top: 50, Bottom: 50, Left: 50, Right: 50}
			}

			buf, err := gen.Generate(tt.tmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

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
	config := DefaultConfig()
	gen, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer gen.Close()

	t.Run("canvas mode from json", func(t *testing.T) {
		json := `{"elements": [{"type": "text", "text": "JSON Canvas Test"}]}`
		buf, err := gen.GenerateFromJSON([]byte(json))
		if err != nil {
			t.Fatalf("GenerateFromJSON() error = %v", err)
		}
		if buf.Len() == 0 {
			t.Error("Generated PDF is empty")
		}
	})

	t.Run("flow mode from json", func(t *testing.T) {
		json := `{
			"mode": "flow",
			"flow_content": [{"height": 20, "columns": [{"size": 12, "text": "JSON Flow Test"}]}]
		}`
		buf, err := gen.GenerateFromJSON([]byte(json))
		if err != nil {
			t.Fatalf("GenerateFromJSON() error = %v", err)
		}
		if buf.Len() == 0 {
			t.Error("Generated PDF is empty")
		}
	})
}
