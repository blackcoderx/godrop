package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
)

func main() {
	// --- PART 1: INPUT VALIDATION ---
	// Need at least 2 args: Program Name, File Name

	downloadLimit := 1 // Limit to 1 download
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <filename>")
		return
	} else if len(os.Args) > 2 {
		limit := os.Args[2]
		downloadLimit, _ = strconv.Atoi(limit)

	}

	targetFile := os.Args[1]

	// Check if file exists
	_, err := os.Stat(targetFile)

	if err != nil {
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
	currentDownloads := 0   // Counter
	var mu sync.Mutex       // The "Key" to protect the counter
	done := make(chan bool) // The "Phone Line" to signal shutdown

	// --- PART 5: HTTP HANDLER ---
	// Handle root URL
	// when the user visits / , the func is called
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// A. LOCK AND CHECK
		mu.Lock() // Lock the counter so that no one else can change it while we check ( preventing race conditions )
		if currentDownloads >= downloadLimit {
			mu.Unlock()
			http.Error(w, "Link Expired", http.StatusGone) // 410 Gone
			return
		}

		// Increment safely
		currentDownloads++
		myNum := currentDownloads
		mu.Unlock() // Unlock ASAP and proceed so that the next person can download the file

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
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Println("Server Error:", err)
			os.Exit(1)
		}
	}()

	// --- PART 7: WAIT ---
	// Block here until we receive the signal from the handler
	// Synchronization without

	<-done
	fmt.Println("Server stopped. Goodbye!")
}
