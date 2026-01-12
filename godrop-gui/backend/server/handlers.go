package server

import (
	"net/http"
)

// RegisterCommonHandlers sets up clipboard handlers shared across modes
func RegisterCommonHandlers(mux *http.ServeMux, getClipboard func() string, setClipboard func(string)) {
	// API: Get Clipboard Data (Polling)
	mux.HandleFunc("/clipboard-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(getClipboard()))
	})

	// API: Set Clipboard Data
	mux.HandleFunc("/clipboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			text := r.FormValue("text")
			setClipboard(text)
			http.Redirect(w, r, "/clipboard", http.StatusSeeOther)
			return
		}

		current := getClipboard()
		w.Write([]byte(GetClipboardTemplate(current)))
	})
}
