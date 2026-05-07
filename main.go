// Package main is the entry point for the Groupie Tracker web application.
// It sets up routes, initializes templates, and starts the HTTP server.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"groupie-tracker/handlers"
)

func main() {
	usage := func() {
		fmt.Println("Usage: go run . [port]")
		fmt.Println("\nStarts the Groupie Tracker web server.")
		fmt.Println("\nArguments:")
		fmt.Printf("  port      The port number to listen on (default: 8080)\n")
		fmt.Println("\nExamples:")
		fmt.Println("  go run .          # Runs on default port 8080")
		fmt.Println("  go run . 3000     # Runs on port 3000")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		usage()
	}

	port := "8080"
	if len(os.Args) == 2 {
		portStr := os.Args[1]
		portInt, err := strconv.Atoi(portStr)
		if err != nil || portInt < 1 || portInt > 65535 {
			usage()
		}
		port = portStr
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// Parse all HTML templates before handling any requests.
	handlers.InitTemplates()

	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images).
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Application routes.
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/artist", handlers.ArtistHandler)
	mux.HandleFunc("/search", handlers.SearchHandler)

	log.Printf("Groupie Tracker server starting on http://localhost%s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
