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

	"github.com/atotto/clipboard"
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

// --- CLIPBOARD LOGIC ---

// GetSystemClipboard returns the current system clipboard text
func (a *App) GetSystemClipboard() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	return text
}

// SetSystemClipboard sets the system clipboard text
func (a *App) SetSystemClipboard(text string) {
	clipboard.WriteAll(text)
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

// GetDefaultSaveDir returns the user's Downloads directory
func (a *App) GetDefaultSaveDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Downloads")
}

// SelectDirectory opens a dialog to select a directory
func (a *App) SelectDirectory() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Save Location",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
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
	a.registerCommonHandlers(mux)

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

// StartReceiveServer starts the Godrop HTTP server in Receive Mode
func (a *App) StartReceiveServer(port string, saveDir string) (ServerResponse, error) {
	a.serverMutex.Lock()
	defer a.serverMutex.Unlock()

	// Stop existing server if running
	if a.server != nil {
		a.StopServer()
	}

	// Ensure save directory exists
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return ServerResponse{}, fmt.Errorf("save directory does not exist")
	}

	mux := http.NewServeMux()
	a.registerCommonHandlers(mux)

	// 1. GET / - The Upload Page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Godrop - Send File</title>
			<style>
				body {
					font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
					background-color: #fdf6e3;
					color: #657b83;
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					height: 100vh;
					margin: 0;
					text-align: center;
				}
				.container {
					background: #eee8d5;
					padding: 40px;
					border-radius: 8px;
					border: 1px solid #93a1a1;
					box-shadow: 0 4px 12px rgba(0,0,0,0.1);
				}
				h1 { color: #cb4b16; margin-bottom: 20px; }
				input[type="file"] { margin: 20px 0; }
				button {
					background-color: #2aa198;
					color: white;
					border: none;
					padding: 12px 24px;
					border-radius: 4px;
					font-size: 1rem;
					cursor: pointer;
					font-weight: bold;
				}
				button:hover { background-color: #268bd2; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Godrop Receiver</h1>
				<form action="/upload" method="post" enctype="multipart/form-data">
					<input type="file" name="file" required>
					<br>
					<button type="submit">Send File</button>
				</form>
			</div>
		</body>
		</html>
		`
		w.Write([]byte(html))
	})

	// 2. POST /upload - Handle File
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Limit upload size (e.g., 2GB or unlimited) - let's say 10GB max just to be safe
		r.Body = http.MaxBytesReader(w, r.Body, 10<<30)

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "File too big or invalid", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create dest file
		dstPath := filepath.Join(saveDir, handler.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Stream copy
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving file content", http.StatusInternalServerError)
			return
		}

		// Notify Frontend
		wailsRuntime.EventsEmit(a.ctx, "file-received", handler.Filename)

		// Success Page
		w.Write([]byte(`
			<h1 style='color:green; font-family:sans-serif; text-align:center;'>File Sent!</h1>
			<script>setTimeout(() => window.location.href='/', 2000)</script>
		`))
	})

	// Setup Server
	ip := GetOutboundIP()
	fullURL := fmt.Sprintf("http://%s:%s", ip, port)

	png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)
	if err != nil {
		return ServerResponse{}, err
	}

	a.server = &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			wailsRuntime.EventsEmit(a.ctx, "server_error", err.Error())
		}
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

// StartClipboardServer starts the Godrop HTTP server in Clipboard Mode
func (a *App) StartClipboardServer(port string) (ServerResponse, error) {
	a.serverMutex.Lock()
	defer a.serverMutex.Unlock()

	// Stop existing server if running
	if a.server != nil {
		a.StopServer()
	}

	mux := http.NewServeMux()
	a.registerCommonHandlers(mux)

	// Redirect root to clipboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/clipboard", http.StatusSeeOther)
	})

	// find IP
	ip := GetOutboundIP()
	fullURL := fmt.Sprintf("http://%s:%s/clipboard", ip, port)

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
		wailsRuntime.EventsEmit(a.ctx, "server_stopped", true)
	}()

	return ServerResponse{
		IP:      ip,
		Port:    port,
		FullURL: fullURL,
		QRCode:  "data:image/png;base64," + toBase64(png),
	}, nil
}

func (a *App) registerCommonHandlers(mux *http.ServeMux) {
	// API: Get Clipboard Data (Polling)
	mux.HandleFunc("/clipboard-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(a.GetSystemClipboard()))
	})

	// API: Set Clipboard Data
	mux.HandleFunc("/clipboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			text := r.FormValue("text")
			a.SetSystemClipboard(text)
			// Redirect back or return success
			http.Redirect(w, r, "/clipboard", http.StatusSeeOther)
			return
		}

		// GET /clipboard UI
		current := a.GetSystemClipboard()
		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Godrop - Clipboard</title>
    <style>
        body { background: #fdf6e3; color: #657b83; font-family: monospace; padding: 20px; display: flex; flex-direction: column; height: 90vh; margin: 0; }
        h2 { text-align: center; color: #cb4b16; }
        textarea { 
            flex: 1; 
            background: #eee8d5; 
            border: 2px solid #93a1a1; 
            padding: 15px; 
            font-family: monospace; 
            font-size: 16px; 
            color: #586e75; 
            border-radius: 4px;
            resize: none;
            outline: none;
        }
        .btn-group { display: flex; gap: 10px; margin-top: 20px; }
        button { 
            flex: 1; 
            padding: 15px; 
            border: none; 
            border-radius: 4px; 
            font-weight: bold; 
            cursor: pointer; 
            font-size: 16px;
            transition: opacity 0.2s;
        }
        button:active { opacity: 0.7; }
        .btn-copy { background: #859900; color: white; }
        .btn-send { background: #2aa198; color: white; }
    </style>
</head>
<body>
    <h2>📋 Shared Clipboard</h2>
    <form action="/clipboard" method="POST" style="display:flex; flex-direction:column; flex:1;">
        <textarea name="text" id="txt">` + current + `</textarea>
        
        <div class="btn-group">
            <button type="button" class="btn-copy" onclick="copyToPhone()">
                COPY TO PHONE
            </button>
            <button type="submit" class="btn-send">
                SEND TO PC
            </button>
        </div>
    </form>
    <script>
        const txt = document.getElementById('txt');
        
        // Polling logic
        setInterval(async () => {
            if (document.activeElement === txt) return; // Don't update if user is typing
            try {
                const r = await fetch('/clipboard-data');
                const data = await r.text();
                if (data !== txt.value) {
                    txt.value = data;
                }
            } catch(e) {}
        }, 2000);

        function copyToPhone() {
            navigator.clipboard.writeText(txt.value);
            const btn = document.querySelector('.btn-copy');
            const old = btn.innerText;
            btn.innerText = "COPIED!";
            setTimeout(() => btn.innerText = old, 1500);
        }
    </script>
</body>
</html>`
		w.Write([]byte(html))
	})
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
