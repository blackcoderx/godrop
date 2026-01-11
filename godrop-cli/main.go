package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
)

type GodropState struct {
	FileName         string
	FilePath         string
	FileSize         int64
	DownloadLimit    int
	CurrentDownloads int
	SecurityCode     string
	StartTime        time.Time
	ExpiryTime       time.Time
	IsTemp           bool       // If true, the file is a temporary zip and should be deleted on exit
	mu               sync.Mutex // The "Key" that ensures only one goroutine touches 'CurrentDownloads' at a time
	done             chan bool
}

func main() {
	// --- PART 1: CLI FLAGS ---
	limit := flag.Int("limit", 1, "Number of downloads allowed before the server stops")
	code := flag.String("code", "", "Optional security code the user must enter on the landing page")
	port := flag.String("port", "8080", "The port the web server will listen on")
	timeout := flag.Duration("timeout", 0, "Time limit for the share (e.g. 10m, 1h). 0 means no timeout.")
	flag.Parse()

	// Get any remaining arguments (these are the file paths)
	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("Usage: godrop [-limit <n>] [-code <code>] [-timeout <duration>] <file1> [file2...]")
		return
	}

	var targetFile string
	var isTemp bool
	var fileName string
	var fileSize int64

	// --- PART 2: FILE PREPARATION ---
	if len(files) > 1 {
		// If multiple files are provided, we bundle them into a single .zip file
		fmt.Println("Packaging multiple files into a temporary archive...")
		tmpZip, err := os.CreateTemp("", "godrop-*.zip")
		if err != nil {
			fmt.Println("Error creating temp zip:", err)
			return
		}

		// Setup the zip writer
		zw := zip.NewWriter(tmpZip)
		for _, f := range files {
			if err := addFileToZip(zw, f); err != nil {
				fmt.Printf("Warning: Failed to add %s to zip: %v\n", f, err)
			}
		}
		zw.Close()
		tmpZip.Close()

		targetFile = tmpZip.Name()
		isTemp = true // Mark as temporary so we clean it up later
		fileName = "godrop-archive.zip"
		fileInfo, _ := os.Stat(targetFile)
		fileSize = fileInfo.Size()
	} else {
		// Single file logic: verify it exists and get its metadata
		targetFile = files[0]
		fileInfo, err := os.Stat(targetFile)
		if err != nil {
			fmt.Printf("Error: File '%s' not found.\n", targetFile)
			return
		}
		fileName = filepath.Base(targetFile)
		fileSize = fileInfo.Size()
	}

	// Initialize our state object
	state := &GodropState{
		FileName:      fileName,
		FilePath:      targetFile,
		FileSize:      fileSize,
		DownloadLimit: *limit,
		SecurityCode:  *code,
		StartTime:     time.Now(),
		IsTemp:        isTemp,
		done:          make(chan bool),
	}

	// --- PART 3: TIMEOUT LOGIC ---
	if *timeout > 0 {
		state.ExpiryTime = state.StartTime.Add(*timeout)
		// Run a background "timer" that shuts down the server when time is up
		go func() {
			time.Sleep(*timeout)
			fmt.Println("\nTimeout reached. Link expired.")
			state.done <- true
		}()
	}

	// --- PART 4: NETWORK & QR CODE ---
	// Get the local IP address so we can generate the correct link
	ip := GetOutboundIP()
	fullURL := fmt.Sprintf("http://%s:%s", ip, *port)

	fmt.Println("----------------------------------------")
	// Generate the QR code for terminal display
	q, _ := qrcode.New(fullURL, qrcode.Medium)
	fmt.Println(q.ToString(false))
	fmt.Printf("Hosting: %s\n", state.FileName)
	fmt.Printf("Downloads Allowed: %d\n", state.DownloadLimit)
	if state.SecurityCode != "" {
		fmt.Printf("Security Code REQUIRED: %s\n", state.SecurityCode)
	}
	if *timeout > 0 {
		fmt.Printf("Expiry Time: %s\n", state.ExpiryTime.Format("15:04:05"))
	}
	fmt.Printf("Share Link: %s\n", fullURL)
	fmt.Println("----------------------------------------")

	// --- PART 5: HTTP ROUTES (The SaaS API) ---

	// 1. Static Files: Serve the 'web' folder (index.html, style.css, script.js)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// 2. API: Stats - Used by the frontend to show download count and file info
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock() // Lock to read current downloads safely
		defer state.mu.Unlock()

		stats := map[string]interface{}{
			"FileName":   state.FileName,
			"FileSize":   state.FileSize,
			"Limit":      state.DownloadLimit,
			"Current":    state.CurrentDownloads,
			"HasCode":    state.SecurityCode != "",
			"StartTime":  state.StartTime.Unix(),
			"ExpiryTime": state.ExpiryTime.Unix(),
		}
		if state.ExpiryTime.IsZero() {
			stats["ExpiryTime"] = 0
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	// 3. API: Verify - Checks if the code entered by the user matches our SecurityCode
	http.HandleFunc("/api/verify", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Code string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		state.mu.Lock()
		success := body.Code == state.SecurityCode
		state.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": success})
	})

	// 4. API: Download - The actual file transfer endpoint
	http.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		// Check limits before allowing download
		if state.CurrentDownloads >= state.DownloadLimit {
			state.mu.Unlock()
			http.Error(w, "Limit Exceeded", http.StatusGone)
			return
		}
		// Check timeout before allowing download
		if !state.ExpiryTime.IsZero() && time.Now().After(state.ExpiryTime) {
			state.mu.Unlock()
			http.Error(w, "Link Expired", http.StatusGone)
			return
		}

		state.CurrentDownloads++
		myNum := state.CurrentDownloads
		state.mu.Unlock()

		fmt.Printf("[%d/%d] Sending file to %s...\n", myNum, state.DownloadLimit, r.RemoteAddr)

		// Set headers to tell the browser this IS a file download
		w.Header().Set("Content-Disposition", "attachment; filename="+state.FileName)
		http.ServeFile(w, r, state.FilePath)

		// If this was the last allowed download, signal shutdown after a short delay
		if myNum >= state.DownloadLimit {
			go func() {
				time.Sleep(2 * time.Second) // Wait for the network packet to clear
				fmt.Println("\nDownload limit reached. System shutting down...")
				state.done <- true
			}()
		}
	})

	// --- PART 6: SERVER LIFECYCLE ---
	server := &http.Server{Addr: ":" + *port}
	go func() {
		fmt.Printf("GODROP Server Live on :%s\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Critical Server Error:", err)
			os.Exit(1)
		}
	}()

	// Wait for the shutdown signal (either from timeout or download limit)
	<-state.done
	fmt.Println("Closing connections...")
	server.Close()

	// Final Cleanup: delete the zip if we created one
	if isTemp {
		os.Remove(targetFile)
	}
	fmt.Println("Goodbye!")
}

// Helper function to add a file from disk into an open zip writer
func addFileToZip(zw *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate // Use standard compression

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file) // Stream the file content into the zip
	return err
}
