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
		html := `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Godrop - Clipboard</title><style>body { background: #fdf6e3; color: #657b83; font-family: monospace; padding: 20px; display: flex; flex-direction: column; height: 90vh; margin: 0; } h2 { text-align: center; color: #cb4b16; } textarea { flex: 1; background: #eee8d5; border: 2px solid #93a1a1; padding: 15px; font-family: monospace; font-size: 16px; color: #586e75; border-radius: 4px; resize: none; outline: none; } .btn-group { display: flex; gap: 10px; margin-top: 20px; } button { flex: 1; padding: 15px; border: none; border-radius: 4px; font-weight: bold; cursor: pointer; font-size: 16px; transition: opacity 0.2s; } button:active { opacity: 0.7; } .btn-copy { background: #859900; color: white; } .btn-send { background: #2aa198; color: white; } </style></head><body><h2>📋 Shared Clipboard</h2><form action="/clipboard" method="POST" style="display:flex; flex-direction:column; flex:1;"><textarea name="text" id="txt">` + current + `</textarea><div class="btn-group"><button type="button" class="btn-copy" onclick="copyToPhone()">COPY TO PHONE</button><button type="submit" class="btn-send">SEND TO PC</button></div></form><script>const txt = document.getElementById('txt'); setInterval(async () => { if (document.activeElement === txt) return; try { const r = await fetch('/clipboard-data'); const data = await r.text(); if (data !== txt.value) { txt.value = data; } } catch(e) {} }, 2000); function copyToPhone() { navigator.clipboard.writeText(txt.value); const btn = document.querySelector('.btn-copy'); const old = btn.innerText; btn.innerText = "COPIED!"; setTimeout(() => btn.innerText = old, 1500); }</script></body></html>`
		w.Write([]byte(html))
	})
}
