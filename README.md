# GoPDF Generator

A comprehensive Go library and REST API for generating PDF files from JSON templates. Built on top of [gopdf](https://github.com/signintech/gopdf) and [Maroto v2](https://github.com/johnfercher/maroto), it supports all major PDF features including images, tables, shapes, custom fonts, RTL languages, and complex Unicode characters.

## Features

- ✅ **JSON Template-Based**: Define PDFs using simple JSON templates
- ✅ **Text & Formatting**: Multiple fonts, styles, colors, alignments
- ✅ **Images**: Support for PNG, JPEG images from file, URL, or embedded data
- ✅ **Tables**: Complex tables with headers, cell spanning, styling
- ✅ **Shapes**: Rectangles, ellipses, lines with fill and stroke colors
- ✅ **Card Element**: Container elements with nested content, shadows, and rounded corners
- ✅ **Enhanced Tables**: Full HTML-table-like control with row span, column span, cell borders
- ✅ **Advanced Layout**: Support for Z-Index (layering) and element Opacity (transparency)
- ✅ **Headers & Footers**: Repeatable template elements on every page
- ✅ **Page Backgrounds**: Define global background colors for PDF pages
- ✅ **Dashed/Dotted Lines**: Apply line styles to borders and shapes
- ✅ **Grid System**: Snap elements to a grid and draw visible grid lines
- ✅ **Watermarks**: Easily add a custom text watermark centered on every page
- ✅ **Custom Fonts**: Load and use TrueType (.ttf) and OpenType (.otf) fonts
- ✅ **RTL Support**: Full support for Arabic, Hebrew, Persian, and other RTL languages
- ✅ **Unicode**: Support for complex Unicode characters and international text
- ✅ **REST API**: HTTP API for remote PDF generation
- ✅ **Library**: Use as a Go package in your applications
- ✅ **Flow Mode**: Dynamic layouts using Maroto v2 with automatic page breaks

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

## Modes of Operation

### Canvas Mode (Default)
Fixed positioning with absolute coordinates. Best for precise layouts like invoices, certificates, and reports.

```json
{
  "mode": "canvas",
  "elements": [
    {
      "type": "text",
      "text": "Hello World",
      "position": {"x": 100, "y": 100}
    }
  ]
}
```

### Flow Mode
Dynamic layouts with automatic page breaks. Best for documents with variable content length.

```json
{
  "mode": "flow",
  "flow_content": [
    {
      "height": 20,
      "columns": [
        {"size": 12, "text": "Dynamic content that flows automatically"}
      ]
    }
  ]
}
```

## Card Element

Card elements are containers that can hold other elements with styling options:

```json
{
  "type": "card",
  "size": {"width": 400, "height": 150},
  "background_color": {"r": 255, "g": 255, "b": 255},
  "border": {"all": true},
  "border_color": {"r": 200, "g": 200, "b": 200},
  "radius": 10,
  "shadow": {
    "color": {"r": 0, "g": 0, "b": 0, "a": 30},
    "offset_x": 3,
    "offset_y": 3,
    "blur": 10
  },
  "padding": {"top": 15, "bottom": 15, "left": 20, "right": 20},
  "elements": [
    {
      "type": "text",
      "text": "Card Title",
      "font": {"size": 18, "style": "B"}
    },
    {
      "type": "newline",
      "height": 10
    },
    {
      "type": "text",
      "text": "Card content goes here..."
    }
  ]
}
```

### Card Properties

| Property | Type | Description |
|----------|------|-------------|
| `size` | Object | Width and height of the card |
| `position` | Object | Absolute position (optional) |
| `background_color` | Object | Background color (R, G, B) |
| `border` | Object | Border configuration |
| `border_color` | Object | Border color (R, G, B) |
| `radius` | Number | Border radius for rounded corners |
| `shadow` | Object | Shadow effect configuration |
| `padding` | Object | Internal padding for content |
| `elements` | Array | Nested child elements |

## Enhanced Table Element

Tables with full HTML-table-like control:

```json
{
  "type": "table",
  "caption": "Sales Report Q4 2024",
  "width": 515,
  "columns": [
    {"width": 120},
    {"width": 80},
    {"width": 80},
    {"width": 80}
  ],
  "header": {
    "cells": [
      {"text": "Region", "font": {"style": "B"}},
      {"text": "Q1", "align": "C"},
      {"text": "Q2", "align": "C"},
      {"text": "Total", "align": "R"}
    ],
    "background": {"r": 52, "g": 73, "b": 94},
    "repeat": true
  },
  "rows": [
    {
      "cells": [
        {"text": "North America", "row_span": 2, "vertical_align": "M"},
        {"text": "$120K", "align": "R"},
        {"text": "$135K", "align": "R"},
        {"text": "$255K", "align": "R", "font": {"style": "B"}}
      ]
    },
    {
      "cells": [
        {"text": "$115K", "align": "R"},
        {"text": "$128K", "align": "R"},
        {"text": "$243K", "align": "R"}
      ]
    }
  ],
  "footer": {
    "cells": [
      {"text": "Grand Total", "col_span": 3},
      {"text": "$1.86M", "align": "R"}
    ]
  },
  "cell_padding": {"top": 8, "bottom": 8, "left": 8, "right": 8},
  "border": {"all": true},
  "min_row_height": 22
}
```

### Table Cell Properties

| Property | Type | Description |
|----------|------|-------------|
| `text` | String | Cell content |
| `font` | Object | Font styling |
| `align` | String | Horizontal alignment (L, C, R) |
| `vertical_align` | String | Vertical alignment (T, M, B) |
| `col_span` | Number | Column spanning |
| `row_span` | Number | Row spanning |
| `background` | Object | Cell background color |
| `border` | Object | Cell-specific borders |
| `border_color` | Object | Cell border color |
| `padding` | Object | Cell-specific padding |
| `elements` | Array | Nested elements in cell |

## Flow Mode Components

Flow mode supports various component types through the enhanced Maroto adapter:

### Text Component
```json
{
  "size": 6,
  "text": "Hello World",
  "style": {
    "font": {"size": 14, "style": "B"},
    "alignment": "center",
    "text_color": {"r": 41, "g": 128, "b": 185},
    "bg_color": {"r": 240, "g": 240, "b": 240}
  }
}
```

### Image Component
```json
{
  "size": 4,
  "image": {
    "path": "./logo.png",
    "width": 100,
    "height": 50
  }
}
```

### QR Code Component
```json
{
  "size": 3,
  "components": [
    {
      "type": "qrcode",
      "text": "https://example.com"
    }
  ]
}
```

### Barcode Component
```json
{
  "size": 4,
  "components": [
    {
      "type": "barcode",
      "text": "123456789012"
    }
  ]
}
```

### Signature Component
```json
{
  "size": 4,
  "components": [
    {
      "type": "signature",
      "text": "John Doe"
    }
  ]
}
```

## RTL (Right-to-Left) Support

For RTL languages like Arabic, Hebrew, Persian:

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
- `card_example.json` - Card element demonstration
- `enhanced_table.json` - HTML-table-like features
- `flow_enhanced.json` - Flow mode with Maroto 2 features
- `rtl_arabic.json` - Arabic/RTL text demonstration

## License

MIT License
