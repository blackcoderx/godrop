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
	RegisterCommonHandlers(mux, core.GetSystemClipboard, core.SetSystemClipboard, core.GetHistory)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(GetReceiveTemplate()))
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
