package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	server           *http.Server
	serverMutex      sync.Mutex
	activeFiles      []string
	isTempArchive    bool
	archivePath      string
	downloadLimit    int
	currentDownloads int
	expiryTime       time.Time
}

// FileEntry represents a file in the explorer
type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"img"` // We use 'img' key to match some frontend expects, or just 'path'. Let's use 'path' and 'type'.
	IsDir bool   `json:"isDir"`
	Size  string `json:"size"`
	Type  string `json:"type"` // "file" or "folder"
}

// ServerResponse returns to the frontend
type ServerResponse struct {
	IP      string `json:"ip"`
	Port    string `json:"port"`
	FullURL string `json:"fullUrl"`
	QRCode  string `json:"qrCode"` // Base64 encoded PNG
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- FILE SYSTEM NAVIGATION ---

// GetHomeDir returns the user's home directory
func (a *App) GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// ReadDir lists files in a directory
func (a *App) ReadDir(path string) ([]FileEntry, error) {
	// Handle "root" for Windows drives
	if (path == "" || path == "root") && runtime.GOOS == "windows" {
		var drives []FileEntry
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drivePath := string(drive) + ":\\"
			_, err := os.Stat(drivePath)
			if err == nil {
				drives = append(drives, FileEntry{
					Name:  drivePath,
					Path:  drivePath,
					IsDir: true,
					Type:  "folder",
					Size:  "",
				})
			}
		}
		return drives, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, e := range entries {
		// Skip hidden files
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}

		info, err := e.Info()
		size := ""
		if err == nil {
			size = formatSize(info.Size())
		}

		entry := FileEntry{
			Name:  e.Name(),
			Path:  filepath.Join(path, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
			Type:  "file",
		}
		if e.IsDir() {
			entry.Type = "folder"
			entry.Size = ""
		}
		files = append(files, entry)
	}
	return files, nil
}

// --- SERVER LOGIC ---

// StartServer starts the Godrop HTTP server
func (a *App) StartServer(port string, password string, files []string, limit int, timeout int) (ServerResponse, error) {
	a.serverMutex.Lock()
	defer a.serverMutex.Unlock()

	// Stop existing server if running
	if a.server != nil {
		a.StopServer()
	}

	if len(files) == 0 {
		return ServerResponse{}, fmt.Errorf("no files selected")
	}

	// Prepare file(s)
	var targetFile string
	var fileName string
	var fileSize int64

	if len(files) > 1 {
		// Zip multiple files
		tmpZip, err := os.CreateTemp("", "godrop-*.zip")
		if err != nil {
			return ServerResponse{}, err
		}
		zw := zip.NewWriter(tmpZip)
		for _, f := range files {
			if err := addFileToZip(zw, f); err != nil {
				log.Printf("Failed to add %s: %v", f, err)
			}
		}
		zw.Close()
		tmpZip.Close()

		targetFile = tmpZip.Name()
		a.isTempArchive = true
		a.archivePath = targetFile
		fileName = "godrop-archive.zip"
	} else {
		// Single file
		targetFile = files[0]
		fileName = filepath.Base(targetFile)
		a.isTempArchive = false
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		return ServerResponse{}, err
	}
	fileSize = info.Size()

	// Initialize Server State
	a.downloadLimit = limit
	a.currentDownloads = 0
	if timeout > 0 {
		a.expiryTime = time.Now().Add(time.Duration(timeout) * time.Minute)
		// Auto-shutdown on timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Minute)
			if a.server != nil {
				wailsRuntime.EventsEmit(a.ctx, "server_error", "Timeout Reached. Server Stopping.")
				a.StopServer()
			}
		}()
	} else {
		a.expiryTime = time.Time{}
	}

	// Setup Server Handler
	mux := http.NewServeMux()

	// API: Stats
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"filename": fileName,
			"size":     fileSize,
			"password": password != "",
		})
	})

	// API: Verify Password
	mux.HandleFunc("/api/verify", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Code string }
		json.NewDecoder(r.Body).Decode(&body)
		authorized := password == "" || body.Code == password
		json.NewEncoder(w).Encode(map[string]bool{"success": authorized})
	})

	// API: Download
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		a.serverMutex.Lock()
		if !a.expiryTime.IsZero() && time.Now().After(a.expiryTime) {
			a.serverMutex.Unlock()
			http.Error(w, "Link Expired", http.StatusGone)
			return
		}
		if a.downloadLimit > 0 && a.currentDownloads >= a.downloadLimit {
			a.serverMutex.Unlock()
			http.Error(w, "Limit Exceeded", http.StatusGone)
			return
		}
		a.currentDownloads++
		current := a.currentDownloads
		a.serverMutex.Unlock()

		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		http.ServeFile(w, r, targetFile)

		wailsRuntime.EventsEmit(a.ctx, "download_started", map[string]string{"ip": r.RemoteAddr})

		// Auto-shutdown if limit reached
		if a.downloadLimit > 0 && current >= a.downloadLimit {
			go func() {
				time.Sleep(2 * time.Second) // allow transfer to start/finish packet
				a.StopServer()
			}()
		}
	})

	// Landing Page (Simple HTML)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body style="font-family:monospace; background:#111; color:#0f0; display:flex; flex-direction:column; align-items:center; justify-content:center; height:100vh;">
		<h1>GODROP DESTINATION</h1>
		<p>File: ` + fileName + `</p>
		<p>Size: ` + formatSize(fileSize) + `</p>
		<p id="msg"></p>
		`
		if password != "" {
			html += `<input type="password" id="pass" placeholder="Password"><button onclick="verify()">Unlock</button>
			<script>
			async function verify() {
				const c = document.getElementById('pass').value;
				const r = await fetch('/api/verify', {method:'POST', body:JSON.stringify({Code:c})});
				const j = await r.json();
				if(j.success) window.location.href='/download';
				else document.getElementById('msg').innerText = "Access Denied";
			}
			</script>`
		} else {
			html += `<button onclick="window.location.href='/download'" style="padding:20px; font-weight:bold; cursor:pointer;">DOWNLOAD NOW</button>`
		}

		html += `</body></html>`
		w.Write([]byte(html))
	})

	// find IP
	ip := GetOutboundIP()
	fullURL := fmt.Sprintf("http://%s:%s", ip, port)

	// Generate QR
	png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)
	if err != nil {
		return ServerResponse{}, err
	}

	// Start Server
	a.server = &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			wailsRuntime.EventsEmit(a.ctx, "server_error", err.Error())
		}
		// Server stopped
		wailsRuntime.EventsEmit(a.ctx, "server_stopped", true)
	}()

	return ServerResponse{
		IP:      ip,
		Port:    port,
		FullURL: fullURL,
		QRCode:  "data:image/png;base64," + toBase64(png),
	}, nil
}

// StopServer stops the HTTP server
func (a *App) StopServer() {
	if a.server != nil {
		a.server.Shutdown(context.Background())
		a.server = nil
	}
	// Cleanup temp files
	if a.isTempArchive && a.archivePath != "" {
		os.Remove(a.archivePath)
		a.isTempArchive = false
	}
}

// --- HELPERS ---

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

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
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// toBase64 helper
func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
