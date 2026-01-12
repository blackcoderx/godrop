package server

import (
	"fmt"
	"net/http"
	"strconv"

	"godrop-gui/backend"

	"github.com/skip2/go-qrcode"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func StartClipboard(core *backend.Core, port string) (ServerResponse, error) {
	mux := http.NewServeMux()
	RegisterCommonHandlers(mux, core.GetSystemClipboard, core.SetSystemClipboard)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/clipboard", http.StatusSeeOther)
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
	fullURL := fmt.Sprintf("http://%s:%s/clipboard", ip, portStr)

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
