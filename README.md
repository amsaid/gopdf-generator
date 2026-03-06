# GoPDF Generator

A comprehensive Go library and REST API for generating PDF files from JSON templates. Built on top of [gopdf](https://github.com/signintech/gopdf), it supports all major PDF features including images, tables, shapes, custom fonts, RTL languages, and complex Unicode characters.

## Features

- ✅ **JSON Template-Based**: Define PDFs using simple JSON templates
- ✅ **Text & Formatting**: Multiple fonts, styles, colors, alignments
- ✅ **Images**: Support for PNG, JPEG images from file, URL, or embedded data
- ✅ **Tables**: Complex tables with headers, cell spanning, styling
- ✅ **Shapes**: Rectangles, ellipses, lines with fill and stroke colors
- ✅ **Advanced Layout**: Support for Z-Index (layering) and element Opacity (transparency)
- ✅ **Headers & Footers**: Repeatable template elements on every page
- ✅ **Page Backgrounds**: Define global background colors for PDF pages
- ✅ **Dashed/Dotted Lines**: Apply line styles to borders and shapes
- ✅ **Grid System**: Snap elements to a grid and draw visible grid lines
- ✅ **Watermarks**: Easily add a custom text watermark centered on every page
- ✅ **Paragraph Indentation**: Easily indent blocks of text or elements via the `indent` property
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
  "background": {
    "r": 250, "g": 250, "b": 250
  },
  "header": [
    {
      "type": "text",
      "text": "Company Header",
      "alignment": {"horizontal": "C"}
    }
  ],
  "footer": [
    {
      "type": "text",
      "text": "Page Footer Content"
    }
  ],
  "watermark": {
    "text": "CONFIDENTIAL",
    "opacity": 0.25,
    "font": {"size": 60, "color": {"r": 255, "g": 0, "b": 0}}
  },
  "grid": {
    "size": 20,
    "draw": true
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
      "alignment": {"horizontal": "C"},
      "z_index": 1,
      "opacity": 0.8,
      "indent": 20
    }
  ]
}
```

## Advanced Document Options

### Headers and Footers
Headers and Footers are arrays of normal `Element` objects (`text`, `image`, `line`, etc.). They are automatically repeated on every page, drawn outside of the standard flow rendering.

### Document Background
You can define a global background color for the entire PDF by providing a `background` color object containing `r`, `g`, and `b` properties at the root level.

### Grid System
Easily implement a grid structure to help align absolute positioned elements by including the `grid` option on your template. Elements will snap to the nearest grid coordinates if `Position` is set. Set `draw: true` to print light dashed guidelines.

### Watermarks
A built-in `watermark` property at the document level creates centered overlay text spanning the document underneath regular text flow but above page background colors.

## Advanced Element Options

### Opacity and Z-Index
You can control the layering and transparency of **any element** using the `z_index` (integer) and `opacity` (float between 0.0 and 1.0) properties. Elements with higher `z_index` values are drawn on top.

```json
{
  "type": "rect",
  "size": {"width": 100, "height": 50},
  "fill_color": {"r": 255, "g": 0, "b": 0},
  "z_index": 5,
  "opacity": 0.5
}
```

### Border and Line Styles
Lines, rectangles, ellipses, and table borders support the `line_style` property. Valid values are `solid` (default), `dashed`, or `dotted`.

```json
{
  "type": "line",
  "end_x": 500,
  "line_width": 2,
  "line_style": "dashed"
}
```

### Indentation
To apply an exact horizontal indentation to a particular element (like a text block or a list), add the `indent: <points>` attribute.

```json
{
  "type": "text",
  "text": "Indented paragraph...",
  "indent": 40
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

## Examples

See the `examples/` directory for complete examples:

- `basic_invoice.json` - Professional invoice template
- `complex_report.json` - Report with tables, shapes, and styling
- `rtl_arabic.json` - Arabic/RTL text demonstration
- `usage_example.go` - Programmatic usage examples

## License

MIT License
