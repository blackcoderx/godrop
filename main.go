package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode" // this is a third-party library for QR code generation
)

func main() {
	// --- PART 1: SETUP ---
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <filename>")
		return
	}
	targetFile := os.Args[1]

	// Check file existence
	if _, err := os.Stat(targetFile); err != nil {
		fmt.Println("File not found")
		return
	}

	ip := GetOutboundIP()
	port := "8080"

	fullURL := fmt.Sprintf("http://%s:%s", ip, port)

	// --- PART 3: QR CODE GENERATION ---
	// Generate a QR code for the URL
	// 'qrcode.Medium' is the error correction level
	q, err := qrcode.New(fullURL, qrcode.Medium)
	if err != nil {
		fmt.Println("Could not generate QR code")
		return
	}

	// Print the QR code to the terminal
	// 'false' usually looks better on black backgrounds
	// 'true' looks better on white backgrounds
	fmt.Println(q.ToString(false))

	fmt.Println("----------------------------------------")
	fmt.Printf("Hosting: %s\n", filepath.Base(targetFile))
	fmt.Printf("Scan above or visit: %s\n", fullURL)
	fmt.Println("----------------------------------------")

	// --- PART 4: SERVER ---
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cleanName := filepath.Base(targetFile)
		w.Header().Set("Content-Disposition", "attachment; filename="+cleanName)
		fmt.Printf("New Connection from: %s\n", r.RemoteAddr)
		http.ServeFile(w, r, targetFile)
	})

	http.ListenAndServe(":"+port, nil)
}
