package parser

import (
	"strings"
	"testing"
)

func TestParseTemplateBytes(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedMode string
		wantErr      bool
	}{
		{
			name: "valid canvas template (default mode)",
			json: `{
				"page_size": "A4",
				"elements": [{"type": "text", "text": "Hello"}]
			}`,
			expectedMode: "canvas",
			wantErr:      false,
		},
		{
			name: "valid flow template",
			json: `{
				"mode": "flow",
				"flow_content": [{"columns": [{"size": 12, "text": "Flow Test"}]}]
			}`,
			expectedMode: "flow",
			wantErr:      false,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
		{
			name: "custom page size",
			json: `{
				"page_width": 500, "page_height": 700,
				"elements": [{"type": "text", "text": "Test"}]
			}`,
			expectedMode: "canvas",
			wantErr:      false,
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

			if template.Mode != tt.expectedMode {
				t.Errorf("Expected mode '%s', got '%s'", tt.expectedMode, template.Mode)
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
			name: "valid canvas template with elements",
			tmpl: DocumentTemplate{
				Mode:     "canvas",
				Elements: []Element{{Type: "text", Text: "Hello"}},
			},
			wantErr: false,
		},
		{
			name: "invalid canvas template (empty)",
			tmpl: DocumentTemplate{
				Mode:     "canvas",
				Elements: []Element{},
			},
			wantErr: true,
		},
		{
			name: "valid flow template",
			tmpl: DocumentTemplate{
				Mode: "flow",
				FlowContent: []FlowRow{
					{Columns: []FlowColumn{{Size: 12, Text: "Content"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid flow template (empty)",
			tmpl: DocumentTemplate{
				Mode:        "flow",
				FlowContent: []FlowRow{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply defaults for tests that don't specify them
			if tt.tmpl.Mode == "" {
				tt.tmpl.Mode = "canvas"
			}
			err := tt.tmpl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseTemplate(t *testing.T) {
	json := `{
		"mode": "flow",
		"title": "Test Document",
		"page_size": "Letter",
		"margin": {"top": 60, "bottom": 60, "left": 60, "right": 60},
		"flow_content": [{"columns": [{"size": 12, "text": "Test content"}]}]
	}`

	reader := strings.NewReader(json)
	template, err := ParseTemplate(reader)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	if template.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", template.Title)
	}
	if template.Mode != "flow" {
		t.Errorf("Expected mode 'flow', got '%s'", template.Mode)
	}
	if len(template.FlowContent) != 1 {
		t.Error("Expected FlowContent to be parsed")
	}
}
