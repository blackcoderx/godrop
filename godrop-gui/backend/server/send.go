package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"godrop-gui/backend"

	"github.com/skip2/go-qrcode"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func StartSend(core *backend.Core, port string, password string, files []string, limit int, timeout int) (ServerResponse, error) {
	if len(files) == 0 {
		return ServerResponse{}, fmt.Errorf("no files selected")
	}

	var targetFile string
	var fileName string

	if len(files) == 1 {
		fi, err := os.Stat(files[0])
		if err == nil && fi.IsDir() {
			var zipErr error
			targetFile, fileName, zipErr = core.CreateZipArchive(files)
			if zipErr != nil {
				return ServerResponse{}, zipErr
			}
		} else {
			targetFile = files[0]
			fileName = filepath.Base(targetFile)
			core.IsTempArchive = false
		}
	} else {
		var zipErr error
		targetFile, fileName, zipErr = core.CreateZipArchive(files)
		if zipErr != nil {
			return ServerResponse{}, zipErr
		}
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		return ServerResponse{}, err
	}
	fileSize := info.Size()

	core.DownloadLimit = limit
	core.CurrentDownloads = 0
	if timeout > 0 {
		core.ExpiryTime = time.Now().Add(time.Duration(timeout) * time.Minute)
		go func() {
			time.Sleep(time.Duration(timeout) * time.Minute)
			core.ServerMutex.Lock()
			defer core.ServerMutex.Unlock()
			if core.Server != nil {
				wailsRuntime.EventsEmit(core.Ctx, "server_error", "Timeout Reached. Server Stopping.")
				Stop(core)
			}
		}()
	} else {
		// Infinity Connection
		core.ExpiryTime = time.Time{}
	}

	mux := http.NewServeMux()
	RegisterCommonHandlers(mux, core.GetSystemClipboard, core.SetSystemClipboard, core.GetHistory)

	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"filename": fileName,
			"size":     fileSize,
			"password": password != "",
		})
	})

	mux.HandleFunc("/api/verify", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Code string }
		json.NewDecoder(r.Body).Decode(&body)
		authorized := password == "" || body.Code == password
		json.NewEncoder(w).Encode(map[string]bool{"success": authorized})
	})

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		core.ServerMutex.Lock()
		if !core.ExpiryTime.IsZero() && time.Now().After(core.ExpiryTime) {
			core.ServerMutex.Unlock()
			http.Error(w, "Link Expired", http.StatusGone)
			return
		}
		if core.DownloadLimit > 0 && core.CurrentDownloads >= core.DownloadLimit {
			core.ServerMutex.Unlock()
			http.Error(w, "Limit Exceeded", http.StatusGone)
			return
		}
		core.CurrentDownloads++
		current := core.CurrentDownloads
		core.ServerMutex.Unlock()

		wailsRuntime.EventsEmit(core.Ctx, "download_started", map[string]string{"ip": r.RemoteAddr})
		pt := &backend.ProgressTracker{Total: fileSize, EventName: "transfer-progress", Ctx: core.Ctx}
		pw := &backend.ProgressResponseWriter{ResponseWriter: w, Tracker: pt}

		ext := strings.ToLower(filepath.Ext(fileName))
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			mimeMap := map[string]string{
				".pdf": "application/pdf", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
				".png": "image/png", ".gif": "image/gif", ".mp4": "video/mp4",
				".zip": "application/zip", ".txt": "text/plain; charset=utf-8",
			}
			contentType = mimeMap[ext]
		}
		if contentType == "" {
			if f, err := os.Open(targetFile); err == nil {
				buffer := make([]byte, 512)
				n, _ := f.Read(buffer)
				contentType = http.DetectContentType(buffer[:n])
				f.Close()
			}
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		encodedName := url.PathEscape(fileName)
		disposition := fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fileName, encodedName)
		w.Header().Set("Content-Disposition", disposition)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		http.ServeFile(pw, r, targetFile)

		if core.DownloadLimit > 0 && current >= core.DownloadLimit {
			go func() { time.Sleep(5 * time.Second); core.ServerMutex.Lock(); Stop(core); core.ServerMutex.Unlock() }()
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(GetSendTemplate(fileName, backend.FormatSize(fileSize), password != "")))
	})

	ip := backend.GetOutboundIP()
	prefPort, _ := strconv.Atoi(port)
	if prefPort == 0 {
		prefPort = 8080
	}
	actualPort, err := backend.FindAvailablePort(prefPort)
	if err != nil {
		return ServerResponse{}, err
	}
	portStr := strconv.Itoa(actualPort)
	fullURL := fmt.Sprintf("http://%s:%s", ip, portStr)

	png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)
	if err != nil {
		return ServerResponse{}, err
	}

	core.Server = &http.Server{Addr: ":" + portStr, Handler: mux}
	go func() {
		if err := core.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			wailsRuntime.EventsEmit(core.Ctx, "server_error", err.Error())
		}
		wailsRuntime.EventsEmit(core.Ctx, "server_stopped", true)
	}()

	return ServerResponse{IP: ip, Port: portStr, FullURL: fullURL, QRCode: "data:image/png;base64," + backend.ToBase64(png)}, nil
}
