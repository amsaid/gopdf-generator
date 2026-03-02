package parser

import (
	"strings"
	"testing"
)

func TestParseTemplateBytes(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid template",
			json: `{
				"page_size": "A4",
				"elements": [
					{"type": "text", "text": "Hello"}
				]
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
		{
			name: "empty elements uses defaults",
			json: `{
				"elements": []
			}`,
			wantErr: false,
		},
		{
			name: "custom page size",
			json: `{
				"page_width": 500,
				"page_height": 700,
				"elements": [{"type": "text", "text": "Test"}]
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := ParseTemplateBytes([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTemplateBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Check defaults were applied
			if template.PageSize == "" && template.PageWidth == 0 {
				t.Error("Expected default page size to be set")
			}
			if template.Orientation == "" {
				t.Error("Expected default orientation to be set")
			}
			if template.Margin == nil {
				t.Error("Expected default margin to be set")
			}
			if template.DefaultFont == nil {
				t.Error("Expected default font to be set")
			}
		})
	}
}

func TestDocumentTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    DocumentTemplate
		wantErr bool
	}{
		{
			name: "valid template with elements",
			tmpl: DocumentTemplate{
				Elements: []Element{
					{Type: "text", Text: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty elements",
			tmpl:    DocumentTemplate{Elements: []Element{}},
			wantErr: true,
		},
		{
			name: "missing element type",
			tmpl: DocumentTemplate{
				Elements: []Element{
					{Text: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid table",
			tmpl: DocumentTemplate{
				Elements: []Element{
					{
						Type: "table",
						Columns: []TableColumn{
							{Width: 100},
							{Width: 100},
						},
						Rows: []TableRow{
							{Cells: []TableCell{{Text: "Cell1"}, {Text: "Cell2"}}},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "table without columns",
			tmpl: DocumentTemplate{
				Elements: []Element{
					{Type: "table", Rows: []TableRow{}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tmpl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseTemplate(t *testing.T) {
	json := `{
		"title": "Test Document",
		"author": "Test Author",
		"page_size": "Letter",
		"orientation": "landscape",
		"margin": {"top": 60, "bottom": 60, "left": 60, "right": 60},
		"default_font": {"family": "Times-Roman", "size": 14},
		"elements": [
			{"type": "text", "text": "Test content"}
		]
	}`

	reader := strings.NewReader(json)
	template, err := ParseTemplate(reader)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	if template.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", template.Title)
	}
	if template.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", template.Author)
	}
	if template.PageSize != "Letter" {
		t.Errorf("Expected page size 'Letter', got '%s'", template.PageSize)
	}
	if template.Orientation != "landscape" {
		t.Errorf("Expected orientation 'landscape', got '%s'", template.Orientation)
	}
	if template.Margin.Top != 60 {
		t.Errorf("Expected top margin 60, got %f", template.Margin.Top)
	}
	if template.DefaultFont.Family != "Times-Roman" {
		t.Errorf("Expected font family 'Times-Roman', got '%s'", template.DefaultFont.Family)
	}
	if template.DefaultFont.Size != 14 {
		t.Errorf("Expected font size 14, got %f", template.DefaultFont.Size)
	}
}

func TestColor(t *testing.T) {
	json := `{
		"elements": [
			{
				"type": "text",
				"text": "Colored text",
				"font": {
					"color": {"r": 255, "g": 128, "b": 0}
				}
			}
		]
	}`

	template, err := ParseTemplateBytes([]byte(json))
	if err != nil {
		t.Fatalf("ParseTemplateBytes() error = %v", err)
	}

	if len(template.Elements) == 0 {
		t.Fatal("Expected at least one element")
	}

	elem := template.Elements[0]
	if elem.Font == nil {
		t.Fatal("Expected font to be set")
	}

	if elem.Font.Color == nil {
		t.Fatal("Expected color to be set")
	}

	if elem.Font.Color.R != 255 || elem.Font.Color.G != 128 || elem.Font.Color.B != 0 {
		t.Errorf("Expected color RGB(255, 128, 0), got RGB(%d, %d, %d)",
			elem.Font.Color.R, elem.Font.Color.G, elem.Font.Color.B)
	}
}

func TestTableStructure(t *testing.T) {
	json := `{
		"elements": [
			{
				"type": "table",
				"columns": [
					{"width": 100, "align": "L"},
					{"width": 150, "align": "C"},
					{"width": 200, "align": "R"}
				],
				"header": {
					"cells": [
						{"text": "Col1", "font": {"style": "B"}},
						{"text": "Col2", "align": "C"},
						{"text": "Col3", "align": "R"}
					],
					"background": {"r": 100, "g": 100, "b": 100},
					"repeat": true
				},
				"rows": [
					{
						"cells": [
							{"text": "Row1Col1", "col_span": 1, "row_span": 1},
							{"text": "Row1Col2", "background": {"r": 200, "g": 200, "b": 200}},
							{"text": "Row1Col3", "rtl": true}
						]
					}
				],
				"cell_padding": {"top": 5, "bottom": 5, "left": 10, "right": 10},
				"border": {"all": true}
			}
		]
	}`

	template, err := ParseTemplateBytes([]byte(json))
	if err != nil {
		t.Fatalf("ParseTemplateBytes() error = %v", err)
	}

	if len(template.Elements) == 0 {
		t.Fatal("Expected at least one element")
	}

	table := template.Elements[0]
	if table.Type != "table" {
		t.Errorf("Expected type 'table', got '%s'", table.Type)
	}

	if len(table.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(table.Columns))
	}

	if table.Header == nil {
		t.Fatal("Expected header to be set")
	}

	if len(table.Header.Cells) != 3 {
		t.Errorf("Expected 3 header cells, got %d", len(table.Header.Cells))
	}

	if len(table.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(table.Rows))
	}

	if table.CellPadding == nil {
		t.Fatal("Expected cell padding to be set")
	}

	if table.Border == nil || !table.Border.All {
		t.Error("Expected border.all to be true")
	}
}

func TestShapeElements(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "rectangle",
			json: `{
				"elements": [{
					"type": "rect",
					"position": {"x": 10, "y": 20},
					"size": {"width": 100, "height": 50},
					"fill_color": {"r": 255, "g": 0, "b": 0},
					"line_color": {"r": 0, "g": 0, "b": 255},
					"line_width": 2
				}]
			}`,
		},
		{
			name: "ellipse",
			json: `{
				"elements": [{
					"type": "ellipse",
					"position": {"x": 100, "y": 100},
					"size": {"width": 80, "height": 60}
				}]
			}`,
		},
		{
			name: "line",
			json: `{
				"elements": [{
					"type": "line",
					"position": {"x": 0, "y": 50},
					"end_x": 500,
					"end_y": 50,
					"line_width": 1,
					"line_color": {"r": 200, "g": 200, "b": 200}
				}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := ParseTemplateBytes([]byte(tt.json))
			if err != nil {
				t.Fatalf("ParseTemplateBytes() error = %v", err)
			}

			if len(template.Elements) == 0 {
				t.Fatal("Expected at least one element")
			}

			elem := template.Elements[0]
			if elem.Position == nil {
				t.Error("Expected position to be set")
			}
			if elem.Size == nil {
				t.Error("Expected size to be set")
			}
		})
	}
}