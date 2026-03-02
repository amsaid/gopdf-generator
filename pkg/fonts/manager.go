package fonts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/signintech/gopdf"
)

// Font styles
const (
	StyleRegular = 0
	StyleBold    = gopdf.Bold
	StyleItalic  = gopdf.Italic
	StyleUnder   = gopdf.Underline
)

// Manager handles font registration and management
type Manager struct {
	fonts      map[string]*FontInfo // Key is "Family|Style"
	mu         sync.RWMutex
	fontDir    string
	embedFonts bool
}

// FontInfo stores information about a registered font
type FontInfo struct {
	Name       string
	FilePath   string
	Family     string
	Style      int
	IsUnicode  bool
	IsRTL      bool
	Data       []byte
}

// NewManager creates a new font manager
func NewManager(fontDir string) *Manager {
	return &Manager{
		fonts:      make(map[string]*FontInfo),
		fontDir:    fontDir,
		embedFonts: true,
	}
}

// Clone creates a shallow copy of the manager
func (m *Manager) Clone() *Manager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newMgr := NewManager(m.fontDir)
	for k, v := range m.fonts {
		newMgr.fonts[k] = v
	}
	return newMgr
}

func getFontKey(family string, style int) string {
	return fmt.Sprintf("%s|%d", strings.ToLower(family), style)
}

func parseStyle(styleStr string) int {
	style := StyleRegular
	styleStr = strings.ToUpper(styleStr)
	if strings.Contains(styleStr, "B") {
		style |= StyleBold
	}
	if strings.Contains(styleStr, "I") {
		style |= StyleItalic
	}
	if strings.Contains(styleStr, "U") {
		style |= StyleUnder
	}
	return style
}

// RegisterFont registers a font from file path
func (m *Manager) RegisterFont(name, filePath string) error {
	return m.RegisterFontWithStyle(name, filePath, StyleRegular)
}

// RegisterFontWithStyle registers a font with a specific style
func (m *Manager) RegisterFontWithStyle(name, filePath string, style int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := getFontKey(name, style)
	if _, exists := m.fonts[key]; exists {
		return nil
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("font file not found: %s", filePath)
	}
	
	m.fonts[key] = &FontInfo{
		Name:     name,
		FilePath: filePath,
		Family:   name,
		Style:    style,
	}
	
	return nil
}

// RegisterFontFromBytes registers a font from byte data
func (m *Manager) RegisterFontFromBytes(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := getFontKey(name, StyleRegular)
	if _, exists := m.fonts[key]; exists {
		return nil
	}
	
	// Save to temp file
	if err := os.MkdirAll(m.fontDir, 0755); err != nil {
		return fmt.Errorf("creating font directory: %w", err)
	}

	tempFile := filepath.Join(m.fontDir, name+".ttf")
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp font file: %w", err)
	}
	
	m.fonts[key] = &FontInfo{
		Name:     name,
		FilePath: tempFile,
		Family:   name,
		Style:    StyleRegular,
		Data:     data,
	}
	
	return nil
}

// RegisterFontFromURL downloads and registers a font from URL
func (m *Manager) RegisterFontFromURL(name, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading font: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading font: status %d", resp.StatusCode)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading font data: %w", err)
	}
	
	return m.RegisterFontFromBytes(name, data)
}

// AddFontsToPDF registers all managed fonts to a PDF instance
func (m *Manager) AddFontsToPDF(pdf *gopdf.GoPdf) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, info := range m.fonts {
		if info.FilePath != "" {
			var err error
			if info.Style == StyleRegular {
				err = pdf.AddTTFFont(info.Name, info.FilePath)
			} else {
				err = pdf.AddTTFFontWithOption(info.Name, info.FilePath, gopdf.TtfOption{Style: info.Style})
			}
			if err != nil {
				return fmt.Errorf("adding font %s (%s): %w", info.Name, info.FilePath, err)
			}
		}
	}
	return nil
}

// RegisterStandardFonts registers built-in PDF fonts (or their TTF equivalents if found)
func (m *Manager) RegisterStandardFonts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Standard families and their file name bases
	families := map[string]string{
		"Helvetica": "Helvetica",
		"Arial":     "arial",
		"Times":     "times",
		"Courier":   "cour",
	}
	
	// Suffixes for styles
	styleSuffixes := []struct {
		suffix string
		style  int
	}{
		{"", StyleRegular},
		{"-Bold", StyleBold},
		{"-BoldItalic", StyleBold | StyleItalic},
		{"-Italic", StyleItalic},
		{"bd", StyleBold},
		{"bi", StyleBold | StyleItalic},
		{"i", StyleItalic},
	}
	
	for formalName, fileName := range families {
		for _, s := range styleSuffixes {
			path := filepath.Join(m.fontDir, fileName+s.suffix+".ttf")
			if _, err := os.Stat(path); err == nil {
				key := getFontKey(formalName, s.style)
				if _, exists := m.fonts[key]; !exists {
					m.fonts[key] = &FontInfo{
						Name: formalName, FilePath: path, Family: formalName, Style: s.style,
					}
				}
			}
		}
	}

	// Add the actual "Standard" names as placeholders
	standards := []string{"Helvetica", "Courier", "Times-Roman", "Symbol", "ZapfDingbats"}
	for _, s := range standards {
		key := getFontKey(s, StyleRegular)
		if _, exists := m.fonts[key]; !exists {
			m.fonts[key] = &FontInfo{Name: s, Family: s, Style: StyleRegular}
		}
	}
}

// SetFont sets the current font on the provided PDF instance
func (m *Manager) SetFont(pdf *gopdf.GoPdf, family string, styleStr string, size float64) error {
	style := parseStyle(styleStr)
	
	m.mu.RLock()
	key := getFontKey(family, style)
	_, exists := m.fonts[key]
	if !exists && style != StyleRegular {
		// Try regular if styled version not found
		key = getFontKey(family, StyleRegular)
		_, exists = m.fonts[key]
	}
	m.mu.RUnlock()
	
	// Even if not in our map, try calling gopdf in case it's a built-in we don't know about
	return pdf.SetFont(family, styleStr, size)
}

// IsRegistered checks if a font is registered
func (m *Manager) IsRegistered(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Check for any style of this family
	prefix := strings.ToLower(name) + "|"
	for k := range m.fonts {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// GetFontInfo returns font information
func (m *Manager) GetFontInfo(name string) (*FontInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return regular style by default
	info, exists := m.fonts[getFontKey(name, StyleRegular)]
	return info, exists
}

// ListFonts returns list of registered fonts
func (m *Manager) ListFonts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	families := make(map[string]bool)
	for _, info := range m.fonts {
		families[info.Family] = true
	}
	
	list := make([]string, 0, len(families))
	for f := range families {
		list = append(list, f)
	}
	return list
}

// LoadUnicodeFont loads a Unicode-capable font
func (m *Manager) LoadUnicodeFont(name, filePath string) error {
	if err := m.RegisterFont(name, filePath); err != nil {
		return err
	}
	
	m.mu.Lock()
	key := getFontKey(name, StyleRegular)
	if info, exists := m.fonts[key]; exists {
		info.IsUnicode = true
	}
	m.mu.Unlock()
	
	return nil
}

// LoadRTLFont loads an RTL-capable font
func (m *Manager) LoadRTLFont(name, filePath string) error {
	if err := m.RegisterFont(name, filePath); err != nil {
		return err
	}
	
	m.mu.Lock()
	key := getFontKey(name, StyleRegular)
	if info, exists := m.fonts[key]; exists {
		info.IsUnicode = true
		info.IsRTL = true
	}
	m.mu.Unlock()
	
	return nil
}

// EnsureFontDir ensures the font directory exists
func EnsureFontDir(dir string) error {
	if dir == "" {
		dir = "./fonts"
	}
	return os.MkdirAll(dir, 0755)
}