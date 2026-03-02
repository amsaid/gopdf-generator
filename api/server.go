package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/amsaid/gopdf-generator/pkg/generator"
	"github.com/amsaid/gopdf-generator/pkg/parser"
	"github.com/gin-gonic/gin"
)

// Server represents the PDF generation API server
type Server struct {
	router    *gin.Engine
	generator *generator.PDFGenerator
	config    *ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	FontDir     string
	TempDir     string
	MaxFileSize int64 // in bytes
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:        "8080",
		FontDir:     "./fonts",
		TempDir:     os.TempDir(),
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// NewServer creates a new API server
func NewServer(config *ServerConfig) (*Server, error) {
	if config == nil {
		config = DefaultServerConfig()
	}
	
	// Create generator
	genConfig := &generator.Config{
		FontDir:    config.FontDir,
		TempDir:    config.TempDir,
		EmbedFonts: true,
	}
	
	gen, err := generator.New(genConfig)
	if err != nil {
		return nil, fmt.Errorf("creating generator: %w", err)
	}
	
	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestLogger())
	
	server := &Server{
		router:    router,
		generator: gen,
		config:    config,
	}
	
	// Setup routes
	server.setupRoutes()
	
	return server, nil
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.handleHealth)
	
	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// PDF generation
		v1.POST("/generate", s.handleGenerate)
		v1.POST("/generate/template", s.handleGenerateFromTemplate)
		v1.POST("/generate/upload", s.handleGenerateFromUpload)
		
		// Font management
		v1.GET("/fonts", s.handleListFonts)
		v1.POST("/fonts/upload", s.handleUploadFont)
		v1.POST("/fonts/register", s.handleRegisterFont)
		
		// Templates
		v1.POST("/templates/validate", s.handleValidateTemplate)
	}
	
	// Static files (for generated PDFs if served)
	s.router.Static("/downloads", "./downloads")
}

// Start starts the server
func (s *Server) Start() error {
	addr := ":" + s.config.Port
	fmt.Printf("PDF Generator API server starting on http://localhost%s\n", addr)
	return s.router.Run(addr)
}

// GetRouter returns the Gin router (for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// GenerateRequest represents a PDF generation request
type GenerateRequest struct {
	Template json.RawMessage `json:"template" binding:"required"`
}

// GenerateResponse represents a PDF generation response
type GenerateResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	PDFData   string `json:"pdf_data,omitempty"` // base64 encoded
	FileURL   string `json:"file_url,omitempty"`
	FileSize  int    `json:"file_size,omitempty"`
}

// handleGenerate generates PDF from JSON template
func (s *Server) handleGenerate(c *gin.Context) {
	var req struct {
		Template map[string]interface{} `json:"template" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Convert template to JSON bytes
	templateData, err := json.Marshal(req.Template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process template: " + err.Error(),
		})
		return
	}
	
	// Generate PDF
	buf, err := s.generator.GenerateFromJSON(templateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "PDF generation failed: " + err.Error(),
		})
		return
	}
	
	// Return PDF
	contentType := c.GetHeader("Accept")
	if contentType == "application/pdf" {
		c.Data(http.StatusOK, "application/pdf", buf.Bytes())
		return
	}
	
	// Return base64 encoded PDF
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"pdf_data": base64.StdEncoding.EncodeToString(buf.Bytes()),
		"file_size": buf.Len(),
	})
}

// handleGenerateFromTemplate generates PDF from template file
func (s *Server) handleGenerateFromTemplate(c *gin.Context) {
	// Get template JSON from request body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to read request body: " + err.Error(),
		})
		return
	}
	
	// Generate PDF
	buf, err := s.generator.GenerateFromJSON(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "PDF generation failed: " + err.Error(),
		})
		return
	}
	
	// Return PDF file
	filename := "generated_" + time.Now().Format("20060102_150405") + ".pdf"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// handleGenerateFromUpload generates PDF from uploaded template file
func (s *Server) handleGenerateFromUpload(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("template")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to get uploaded file: " + err.Error(),
		})
		return
	}
	
	// Check file size
	if file.Size > s.config.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("File too large. Max size: %d bytes", s.config.MaxFileSize),
		})
		return
	}
	
	// Open file
	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to open uploaded file: " + err.Error(),
		})
		return
	}
	defer openedFile.Close()
	
	// Generate PDF
	buf, err := s.generator.GenerateFromReader(openedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "PDF generation failed: " + err.Error(),
		})
		return
	}
	
	// Return PDF file
	filename := file.Filename[:len(file.Filename)-len(filepath.Ext(file.Filename))] + ".pdf"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// handleListFonts returns list of registered fonts
func (s *Server) handleListFonts(c *gin.Context) {
	fonts := s.generator.GetFontManager().ListFonts()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"fonts":   fonts,
		"count":   len(fonts),
	})
}

// handleUploadFont handles font file upload
func (s *Server) handleUploadFont(c *gin.Context) {
	// Get font name
	fontName := c.PostForm("name")
	if fontName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Font name is required",
		})
		return
	}
	
	// Get uploaded file
	file, err := c.FormFile("font")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to get uploaded file: " + err.Error(),
		})
		return
	}
	
	// Check file extension
	ext := filepath.Ext(file.Filename)
	if ext != ".ttf" && ext != ".otf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid font file. Only .ttf and .otf files are supported",
		})
		return
	}
	
	// Save font file
	fontPath := filepath.Join(s.config.FontDir, fontName+ext)
	if err := c.SaveUploadedFile(file, fontPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save font file: " + err.Error(),
		})
		return
	}
	
	// Register font
	if err := s.generator.RegisterFont(fontName, fontPath); err != nil {
		os.Remove(fontPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to register font: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Font uploaded and registered successfully",
		"font":    fontName,
		"path":    fontPath,
	})
}

// RegisterFontRequest represents a font registration request
type RegisterFontRequest struct {
	Name     string `json:"name" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
	IsRTL    bool   `json:"is_rtl,omitempty"`
}

// handleRegisterFont registers a font from file path
func (s *Server) handleRegisterFont(c *gin.Context) {
	var req RegisterFontRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Register font
	var err error
	if req.IsRTL {
		err = s.generator.LoadRTLFont(req.Name, req.FilePath)
	} else {
		err = s.generator.RegisterFont(req.Name, req.FilePath)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to register font: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Font registered successfully",
		"font":  req.Name,
	})
}

// handleValidateTemplate validates a template without generating PDF
func (s *Server) handleValidateTemplate(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to read request body: " + err.Error(),
		})
		return
	}
	
	template, err := parser.ParseTemplateBytes(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid template: " + err.Error(),
		})
		return
	}
	
	if err := template.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Template validation failed: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Template is valid",
		"elements_count": len(template.Elements),
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// requestLogger logs HTTP requests
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		fmt.Printf("[PDF-API] %v | %3d | %13v | %15s | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}