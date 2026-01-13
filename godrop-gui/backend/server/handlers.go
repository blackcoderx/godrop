package server

import (
	"net/http"
	"strings"
)

// RegisterCommonHandlers sets up clipboard handlers shared across modes
func RegisterCommonHandlers(mux *http.ServeMux, getClipboard func() string, setClipboard func(string), getHistory func() []string) {
	// API: Get Clipboard Data (Polling)
	mux.HandleFunc("/clipboard-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(getClipboard()))
	})

	// API: Get Full History
	mux.HandleFunc("/clipboard-history", func(w http.ResponseWriter, r *http.Request) {
		history := getHistory()
		w.Header().Set("Content-Type", "application/json")
		// Simple JSON join for history if not wanting full json encoder
		w.Write([]byte(`["` + strings.Join(history, `","`) + `"]`))
	})

	// API: Set Clipboard Data
	mux.HandleFunc("/clipboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			text := r.FormValue("text")
			setClipboard(text)
			http.Redirect(w, r, "/clipboard", http.StatusSeeOther)
			return
		}

		w.Write([]byte(GetClipboardTemplate(getHistory())))
	})
}
