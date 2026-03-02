# GoPDF Generator

A comprehensive Go library and REST API for generating PDF files from JSON templates. Built on top of [gopdf](https://github.com/signintech/gopdf), it supports all major PDF features including images, tables, shapes, custom fonts, RTL languages, and complex Unicode characters.

## Features

- ✅ **JSON Template-Based**: Define PDFs using simple JSON templates
- ✅ **Text & Formatting**: Multiple fonts, styles, colors, alignments
- ✅ **Images**: Support for PNG, JPEG images from file, URL, or embedded data
- ✅ **Tables**: Complex tables with headers, cell spanning, styling
- ✅ **Shapes**: Rectangles, ellipses, lines with fill and stroke colors
- ✅ **Custom Fonts**: Load and use TrueType (.ttf) and OpenType (.otf) fonts
- ✅ **RTL Support**: Full support for Arabic, Hebrew, Persian, and other RTL languages
- ✅ **Unicode**: Support for complex Unicode characters and international text
- ✅ **REST API**: HTTP API for remote PDF generation
- ✅ **Library**: Use as a Go package in your applications

## Installation

```bash
go get github.com/amsaid/gopdf-generator
```

## Quick Start

### As a Library

```go
package main

import (
    "log"
    "os"
    "github.com/amsaid/gopdf-generator/pkg/generator"
)

func main() {
    // Create generator
    gen, err := generator.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer gen.Close()

    // Read template
    templateData, err := os.ReadFile("template.json")
    if err != nil {
        log.Fatal(err)
    }

    // Generate PDF
    buf, err := gen.GenerateFromJSON(templateData)
    if err != nil {
        log.Fatal(err)
    }

    // Save to file
    os.WriteFile("output.pdf", buf.Bytes(), 0644)
}
```

### As a REST API Server

```bash
# Start the server
go run cmd/server/main.go

# Or with custom options
go run cmd/server/main.go -port=8080 -font-dir=./fonts

# Generate PDF via API
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d @template.json \
  --output output.pdf
```

## JSON Template Structure

```json
{
  "title": "Document Title",
  "author": "Author Name",
  "page_size": "A4",
  "orientation": "portrait",
  "margin": {
    "top": 50,
    "bottom": 50,
    "left": 50,
    "right": 50
  },
  "default_font": {
    "family": "Helvetica",
    "size": 12
  },
  "fonts": [
    {
      "name": "CustomFont",
      "file_path": "./fonts/CustomFont.ttf"
    }
  ],
  "elements": [
    {
      "type": "text",
      "text": "Hello World",
      "font": {
        "family": "Helvetica",
        "size": 24,
        "style": "B",
        "color": {"r": 41, "g": 128, "b": 185}
      },
      "alignment": {"horizontal": "C"}
    }
  ]
}
```

## Element Types

### Text
```json
{
  "type": "text",
  "text": "Your text here",
  "font": {
    "family": "Helvetica",
    "size": 12,
    "style": "B",
    "color": {"r": 0, "g": 0, "b": 0}
  },
  "alignment": {"horizontal": "L"},
  "line_height": 18,
  "rtl": false
}
```

### Cell (Text with background/border)
```json
{
  "type": "cell",
  "position": {"x": 50, "y": 100},
  "size": {"width": 200, "height": 50},
  "text": "Cell content",
  "background_color": {"r": 240, "g": 240, "b": 240},
  "border": {"all": true},
  "alignment": {"horizontal": "C", "vertical": "M"}
}
```

### Image
```json
{
  "type": "image",
  "position": {"x": 50, "y": 100},
  "size": {"width": 200, "height": 150},
  "image_path": "./image.png"
}
```

### Table
```json
{
  "type": "table",
  "columns": [
    {"width": 200},
    {"width": 100},
    {"width": 100}
  ],
  "header": {
    "cells": [
      {"text": "Column 1", "font": {"style": "B"}},
      {"text": "Column 2", "align": "C"},
      {"text": "Column 3", "align": "R"}
    ],
    "background": {"r": 52, "g": 73, "b": 94}
  },
  "rows": [
    {
      "cells": [
        {"text": "Row 1 Col 1"},
        {"text": "Row 1 Col 2", "align": "C"},
        {"text": "Row 1 Col 3", "align": "R"}
      ]
    }
  ],
  "cell_padding": {"top": 8, "bottom": 8, "left": 10, "right": 10},
  "border": {"all": true}
}
```

### Rectangle
```json
{
  "type": "rect",
  "position": {"x": 50, "y": 100},
  "size": {"width": 200, "height": 100},
  "fill_color": {"r": 46, "g": 204, "b": 113},
  "line_color": {"r": 39, "g": 174, "b": 96},
  "line_width": 2
}
```

### Ellipse
```json
{
  "type": "ellipse",
  "position": {"x": 200, "y": 200},
  "size": {"width": 100, "height": 60},
  "fill_color": {"r": 155, "g": 89, "b": 182},
  "line_width": 1
}
```

### Line
```json
{
  "type": "line",
  "position": {"x": 50, "y": 100},
  "end_x": 500,
  "end_y": 100,
  "line_width": 1,
  "line_color": {"r": 200, "g": 200, "b": 200}
}
```

### Newline
```json
{
  "type": "newline",
  "height": 20
}
```

### Page Break
```json
{
  "type": "pagebreak"
}
```

## RTL (Right-to-Left) Support

For RTL languages like Arabic, Hebrew, Persian:

1. Use a Unicode font that supports your language (e.g., Noto Sans Arabic)
2. Set `rtl: true` on text elements or table cells
3. The text will be automatically processed for proper RTL display

```json
{
  "fonts": [
    {
      "name": "NotoArabic",
      "file_path": "./fonts/NotoSansArabic-Regular.ttf"
    }
  ],
  "elements": [
    {
      "type": "text",
      "text": "مرحبا بالعالم",
      "font": {"family": "NotoArabic", "size": 18},
      "rtl": true,
      "alignment": {"horizontal": "R"}
    }
  ]
}
```

## REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/generate` | Generate PDF from JSON template |
| POST | `/api/v1/generate/template` | Generate PDF (returns file) |
| POST | `/api/v1/generate/upload` | Generate from uploaded template file |
| GET | `/api/v1/fonts` | List registered fonts |
| POST | `/api/v1/fonts/upload` | Upload and register a font |
| POST | `/api/v1/fonts/register` | Register a font from path |
| POST | `/api/v1/templates/validate` | Validate template without generating |

### API Examples

**Generate PDF:**
```bash
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "template": {
      "page_size": "A4",
      "elements": [
        {"type": "text", "text": "Hello API!", "font": {"size": 24}}
      ]
    }
  }'
```

**Upload Font:**
```bash
curl -X POST http://localhost:8080/api/v1/fonts/upload \
  -F "name=MyFont" \
  -F "font=@/path/to/font.ttf"
```

**Validate Template:**
```bash
curl -X POST http://localhost:8080/api/v1/templates/validate \
  -H "Content-Type: application/json" \
  -d @template.json
```

## Standard Fonts

The following standard PDF fonts are always available:

- **Helvetica** (Normal, Bold, Oblique, BoldOblique)
- **Times-Roman** (Normal, Bold, Italic, BoldItalic)
- **Courier** (Normal, Bold, Oblique, BoldOblique)
- **Symbol**
- **ZapfDingbats**

## Page Sizes

Supported page sizes:
- `A4` (default)
- `A3`
- `A5`
- `Letter`
- `Legal`
- Custom (specify `page_width` and `page_height` in points)

## Configuration Options

### Server Configuration

```go
config := &api.ServerConfig{
    Port:        "8080",
    FontDir:     "./fonts",
    TempDir:     os.TempDir(),
    MaxFileSize: 10 * 1024 * 1024, // 10MB
}

server, err := api.NewServer(config)
```

### Generator Configuration

```go
config := &generator.Config{
    FontDir:    "./fonts",
    TempDir:    os.TempDir(),
    EmbedFonts: true,
}

gen, err := generator.New(config)
```

## Examples

See the `examples/` directory for complete examples:

- `basic_invoice.json` - Professional invoice template
- `complex_report.json` - Report with tables, shapes, and styling
- `rtl_arabic.json` - Arabic/RTL text demonstration
- `usage_example.go` - Programmatic usage examples

## Project Structure

```
gopdf-generator/
├── cmd/
│   └── server/         # REST API server entry point
│       └── main.go
├── pkg/
│   ├── generator/      # Core PDF generation engine
│   │   └── generator.go
│   ├── parser/         # JSON template parser
│   │   └── template.go
│   ├── elements/       # Element renderers
│   │   └── handlers.go
│   ├── fonts/          # Font management
│   │   └── manager.go
│   └── rtl/            # RTL text processing
│       └── handler.go
├── api/                # REST API implementation
│   └── server.go
├── examples/           # Example templates and code
│   ├── basic_invoice.json
│   ├── complex_report.json
│   ├── rtl_arabic.json
│   └── usage_example.go
├── fonts/              # Font storage directory
├── go.mod
└── README.md
```

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [signintech/gopdf](https://github.com/signintech/gopdf) - PDF generation library

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
