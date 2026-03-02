package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/amsaid/gopdf-generator/api"
)

func main() {
	var (
		port    = flag.String("port", getEnv("PORT", "8080"), "Server port")
		fontDir = flag.String("font-dir", getEnv("FONT_DIR", "./fonts"), "Font directory path")
		tempDir = flag.String("temp-dir", getEnv("TEMP_DIR", os.TempDir()), "Temp directory path")
	)
	flag.Parse()

	// Ensure directories exist
	os.MkdirAll(*fontDir, 0755)
	os.MkdirAll(*tempDir, 0755)
	os.MkdirAll("./downloads", 0755)

	// Create server config
	config := &api.ServerConfig{
		Port:    *port,
		FontDir: *fontDir,
		TempDir: *tempDir,
	}

	// Create server
	server, err := api.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		os.Exit(0)
	}()

	// Print startup info
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           GoPDF Generator API Server                       ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Port:     %s\n", *port)
	fmt.Printf("║  Font Dir: %s\n", *fontDir)
	fmt.Printf("║  Temp Dir: %s\n", *tempDir)
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Endpoints:                                                ║")
	fmt.Println("║    GET  /health              - Health check                ║")
	fmt.Println("║    POST /api/v1/generate     - Generate PDF from JSON      ║")
	fmt.Println("║    POST /api/v1/generate/template - Generate from template ║")
	fmt.Println("║    POST /api/v1/generate/upload   - Generate from upload   ║")
	fmt.Println("║    GET  /api/v1/fonts        - List fonts                  ║")
	fmt.Println("║    POST /api/v1/fonts/upload - Upload font file            ║")
	fmt.Println("║    POST /api/v1/fonts/register  - Register font            ║")
	fmt.Println("║    POST /api/v1/templates/validate - Validate template     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
