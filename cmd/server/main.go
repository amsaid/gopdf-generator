package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amsaid/gopdf-generator/api"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

func clr(color, text string) string {
	return color + text + reset
}

func printBanner() {
	fmt.Println()
	fmt.Println(clr(cyan+bold, "  ██████╗  ██████╗ ██████╗ ██████╗ ███████╗"))
	fmt.Println(clr(cyan+bold, " ██╔════╝ ██╔═══██╗██╔══██╗██╔══██╗██╔════╝"))
	fmt.Println(clr(cyan+bold, " ██║  ███╗██║   ██║██████╔╝██║  ██║█████╗  "))
	fmt.Println(clr(cyan+bold, " ██║   ██║██║   ██║██╔═══╝ ██║  ██║██╔══╝  "))
	fmt.Println(clr(cyan+bold, " ╚██████╔╝╚██████╔╝██║     ██████╔╝██╗"))
	fmt.Println(clr(cyan+bold, "  ╚═════╝  ╚═════╝ ╚═╝     ╚═════╝ ╚═╝"))
	fmt.Println()
	fmt.Println(clr(white+bold, "  PDF Generator API") + clr(dim, "  •  Production Ready  •  v1.0.0"))
	fmt.Println()
}

func printDivider() {
	fmt.Println(clr(dim, "  ──────────────────────────────────────────────────────"))
}

func printSection(title string) {
	fmt.Println()
	fmt.Println(clr(yellow+bold, "  ▸ "+title))
	printDivider()
}

func printConfigRow(icon, key, value string) {
	fmt.Printf("  %s  %-14s %s\n",
		icon,
		clr(dim, key),
		clr(white+bold, value),
	)
}

type endpoint struct {
	method string
	path   string
	desc   string
}

func printEndpoint(e endpoint) {
	var methodColor string
	switch e.method {
	case "GET":
		methodColor = green + bold
	case "POST":
		methodColor = blue + bold
	case "DELETE":
		methodColor = red + bold
	default:
		methodColor = white
	}

	fmt.Printf("  %s  %-44s %s\n",
		clr(methodColor, fmt.Sprintf("%-6s", e.method)),
		clr(cyan, e.path),
		clr(dim, e.desc),
	)
}

func printStartupInfo(port, fontDir, tempDir string) {
	printBanner()

	fmt.Println(clr(cyan, "  ══════════════════════════════════════════════════════════"))

	printSection("Configuration")
	printConfigRow("🔌", "Port:", ":"+port)
	printConfigRow("🔤", "Font Dir:", fontDir)
	printConfigRow("📁", "Temp Dir:", tempDir)
	printConfigRow("🕐", "Started:", time.Now().Format("2006-01-02 15:04:05"))

	printSection("Endpoints")

	endpoints := []endpoint{
		{"GET", "/health", "Health check & status"},
		{"POST", "/api/v1/generate", "Generate PDF from JSON"},
		{"POST", "/api/v1/generate/template", "Generate PDF from template"},
		{"POST", "/api/v1/generate/upload", "Generate PDF from file upload"},
		{"GET", "/api/v1/fonts", "List available fonts"},
		{"POST", "/api/v1/fonts/upload", "Upload a font file"},
		{"POST", "/api/v1/fonts/register", "Register an existing font"},
		{"POST", "/api/v1/templates/validate", "Validate a template"},
	}

	for _, e := range endpoints {
		printEndpoint(e)
	}

	fmt.Println()
	fmt.Println(clr(cyan, "  ══════════════════════════════════════════════════════════"))
	fmt.Println()
	fmt.Printf("  %s  Server listening on %s\n",
		clr(green+bold, "✓"),
		clr(cyan+bold, "http://localhost:"+port),
	)
	fmt.Printf("  %s  Press %s to stop\n\n",
		clr(dim, "i"),
		clr(white+bold, "Ctrl+C"),
	)
}

func main() {
	var (
		port    = flag.String("port", getEnv("PORT", "8080"), "Server port")
		fontDir = flag.String("font-dir", getEnv("FONT_DIR", "./fonts"), "Font directory path")
		tempDir = flag.String("temp-dir", getEnv("TEMP_DIR", os.TempDir()), "Temp directory path")
	)
	flag.Parse()

	os.MkdirAll(*fontDir, 0755)
	os.MkdirAll(*tempDir, 0755)
	os.MkdirAll("./downloads", 0755)

	config := &api.ServerConfig{
		Port:    *port,
		FontDir: *fontDir,
		TempDir: *tempDir,
	}

	server, err := api.NewServer(config)
	if err != nil {
		fmt.Printf("\n  %s  Failed to create server: %v\n\n",
			clr(red+bold, "✗"), err)
		log.Fatalf("Failed to create server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		fmt.Printf("  %s  Gracefully shutting down...\n", clr(yellow+bold, "⏹"))
		fmt.Printf("  %s  Goodbye!\n\n", clr(cyan, "👋"))
		os.Exit(0)
	}()

	printStartupInfo(*port, *fontDir, *tempDir)

	if err := server.Start(); err != nil {
		fmt.Printf("\n  %s  Server error: %v\n\n", clr(red+bold, "✗"), err)
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
