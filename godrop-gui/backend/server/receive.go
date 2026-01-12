package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"godrop-gui/backend"

	"github.com/skip2/go-qrcode"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func StartReceive(core *backend.Core, port string, saveDir string) (ServerResponse, error) {
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return ServerResponse{}, fmt.Errorf("save directory does not exist")
	}

	mux := http.NewServeMux()
	RegisterCommonHandlers(mux, core.GetSystemClipboard, core.SetSystemClipboard)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Godrop - Send File</title><style>body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #fdf6e3; color: #657b83; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; margin: 0; text-align: center; } .container { background: #eee8d5; padding: 40px; border-radius: 8px; border: 1px solid #93a1a1; box-shadow: 0 4px 12px rgba(0,0,0,0.1); } h1 { color: #cb4b16; margin-bottom: 20px; } input[type="file"] { margin: 20px 0; } button { background-color: #2aa198; color: white; border: none; padding: 12px 24px; border-radius: 4px; font-size: 1rem; cursor: pointer; font-weight: bold; } button:hover { background-color: #268bd2; } </style></head><body><div class="container"><h1>Godrop Receiver</h1><form action="/upload" method="post" enctype="multipart/form-data"><input type="file" name="file" required><br><button type="submit">Send File</button></form></div></body></html>`
		w.Write([]byte(html))
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
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

		dstPath := filepath.Join(saveDir, handler.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		pt := &backend.ProgressTracker{Total: handler.Size, EventName: "transfer-progress", Ctx: core.Ctx, Reader: file}
		if _, err := io.Copy(dst, pt); err != nil {
			http.Error(w, "Error saving file content", http.StatusInternalServerError)
			return
		}

		wailsRuntime.EventsEmit(core.Ctx, "file-received", handler.Filename)
		w.Write([]byte(`<h1 style='color:green; font-family:sans-serif; text-align:center;'>File Sent!</h1><script>setTimeout(() => window.location.href='/', 2000)</script>`))
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
