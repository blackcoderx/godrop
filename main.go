package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
)

func main() {
	// --- PART 1: INPUT VALIDATION ---
	// Need at least 2 args: Program Name, File Name
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <filename>")
		return
	}
	targetFile := os.Args[1]

	// Check if file exists
	if _, err := os.Stat(targetFile); err != nil {
		fmt.Printf("Error: File '%s' not found.\n", targetFile)
		return
	}

	// --- PART 2: NETWORK SETUP ---
	// Get local IP (requires utils.go)
	ip := GetOutboundIP()
	port := "8080"
	fullURL := fmt.Sprintf("http://%s:%s", ip, port)

	// --- PART 3: QR CODE ---
	fmt.Println("----------------------------------------")
	q, _ := qrcode.New(fullURL, qrcode.Medium)
	fmt.Println(q.ToString(false))
	fmt.Printf("Hosting: %s\n", filepath.Base(targetFile))
	fmt.Printf("Scan to download: %s\n", fullURL)
	fmt.Println("----------------------------------------")

	// --- PART 4: CONCURRENCY STATE ---
	downloadLimit := 1      // Limit to 1 download
	currentDownloads := 0   // Counter
	var mu sync.Mutex       // The "Key" to protect the counter
	done := make(chan bool) // The "Phone Line" to signal shutdown

	// --- PART 5: HTTP HANDLER ---
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// A. LOCK AND CHECK
		mu.Lock()
		if currentDownloads >= downloadLimit {
			mu.Unlock()
			http.Error(w, "Link Expired", http.StatusGone) // 410 Gone
			return
		}

		// Increment safely
		currentDownloads++
		myNum := currentDownloads
		mu.Unlock()

		// B. SERVE FILE
		fmt.Printf("[%d/%d] Transferring to %s...\n", myNum, downloadLimit, r.RemoteAddr)

		// Set headers to force download
		cleanName := filepath.Base(targetFile)
		w.Header().Set("Content-Disposition", "attachment; filename="+cleanName)
		http.ServeFile(w, r, targetFile)

		// C. CHECK FOR SHUTDOWN
		// Run in background so we don't block the file transfer
		if myNum >= downloadLimit {
			go func() {
				// Wait 2 seconds for transfer to finish
				time.Sleep(2 * time.Second)
				fmt.Println("\nLimit reached. Sending shutdown signal...")
				done <- true
			}()
		}
	})

	// --- PART 6: START SERVER ---
	// Run in a Goroutine so main() can keep moving
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Println("Server Error:", err)
			os.Exit(1)
		}
	}()

	// --- PART 7: WAIT ---
	// Block here until we receive the signal from the handler
	<-done
	fmt.Println("Server stopped. Goodbye!")
}
