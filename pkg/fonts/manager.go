package fonts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/signintech/gopdf"
)

// Manager handles font registration and management
type Manager struct {
	pdf        *gopdf.GoPdf
	fonts      map[string]*FontInfo
	mu         sync.RWMutex
	fontDir    string
	embedFonts bool
}

// FontInfo stores information about a registered font
type FontInfo struct {
	Name       string
	FilePath   string
	Family     string
	Style      string
	IsUnicode  bool
	IsRTL      bool
	Data       []byte
}

// NewManager creates a new font manager
func NewManager(pdf *gopdf.GoPdf, fontDir string) *Manager {
	return &Manager{
		pdf:        pdf,
		fonts:      make(map[string]*FontInfo),
		fontDir:    fontDir,
		embedFonts: true,
	}
}

// RegisterFont registers a font from file path
func (m *Manager) RegisterFont(name, filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.fonts[name]; exists {
		return nil // Already registered
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("font file not found: %s", filePath)
	}
	
	// Register with gopdf
	err := m.pdf.AddTTFFont(name, filePath)
	if err != nil {
		return fmt.Errorf("adding TTF font %s: %w", name, err)
	}
	
	m.fonts[name] = &FontInfo{
		Name:     name,
		FilePath: filePath,
		Family:   name,
	}
	
	return nil
}

// RegisterFontFromBytes registers a font from byte data
func (m *Manager) RegisterFontFromBytes(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.fonts[name]; exists {
		return nil // Already registered
	}
	
	// Save to temp file for gopdf
	tempFile := filepath.Join(m.fontDir, name+".ttf")
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp font file: %w", err)
	}
	
	// Register with gopdf
	err := m.pdf.AddTTFFont(name, tempFile)
	if err != nil {
		return fmt.Errorf("adding TTF font %s: %w", name, err)
	}
	
	m.fonts[name] = &FontInfo{
		Name:     name,
		FilePath: tempFile,
		Family:   name,
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

// RegisterStandardFonts registers built-in PDF fonts
func (m *Manager) RegisterStandardFonts() {
	// Standard PDF fonts are always available in gopdf
	m.mu.Lock()
	defer m.mu.Unlock()
	
	standardFonts := []string{
		"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
		"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique",
		"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
		"Symbol", "ZapfDingbats",
	}
	
	for _, font := range standardFonts {
		m.fonts[font] = &FontInfo{
			Name:      font,
			Family:    font,
			IsUnicode: false,
		}
	}
}

// SetFont sets the current font
func (m *Manager) SetFont(family string, style string, size float64) error {
	m.mu.RLock()
	_, exists := m.fonts[family]
	m.mu.RUnlock()
	
	if !exists {
		// Try to use as standard font
		return m.pdf.SetFont(family, style, size)
	}
	
	return m.pdf.SetFont(family, style, size)
}

// IsRegistered checks if a font is registered
func (m *Manager) IsRegistered(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.fonts[name]
	return exists
}

// GetFontInfo returns font information
func (m *Manager) GetFontInfo(name string) (*FontInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, exists := m.fonts[name]
	return info, exists
}

// ListFonts returns list of registered fonts
func (m *Manager) ListFonts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	fonts := make([]string, 0, len(m.fonts))
	for name := range m.fonts {
		fonts = append(fonts, name)
	}
	return fonts
}

// LoadUnicodeFont loads a Unicode-capable font (like Noto Sans)
func (m *Manager) LoadUnicodeFont(name, filePath string) error {
	if err := m.RegisterFont(name, filePath); err != nil {
		return err
	}
	
	m.mu.Lock()
	if info, exists := m.fonts[name]; exists {
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
	if info, exists := m.fonts[name]; exists {
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