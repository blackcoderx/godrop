package server

import (
	"fmt"
	"strings"
)

// Vibrant Retro Design System
const sharedCSS = `
:root {
  --bg: #F7F4F3;
  --bg-card: #FFFFFF;
  --accent: #52a1d8;
  --accent-bright: #FF5A5F;
  --text: #191610;
  --text-muted: #6B6B6B;
  --border: #191610;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
  background: var(--bg);
  color: var(--text);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
}

.retro-container {
  background: var(--bg-card);
  border: 3px solid var(--border);
  border-radius: 30px;
  box-shadow: 12px 12px 0 var(--border);
  width: 100%;
  max-width: 450px;
  padding: 40px;
  text-align: center;
  animation: slideUp 0.4s ease-out;
}

h1 { font-size: 1.5rem; font-weight: 800; letter-spacing: 2px; margin-bottom: 20px; }
p { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 25px; line-height: 1.5; }

.info-card {
  background: #F0F0F0;
  border: 2px solid var(--border);
  padding: 20px;
  border-radius: 15px;
  margin-bottom: 30px;
  text-align: left;
}

.info-item { display: flex; flex-direction: column; gap: 4px; margin-bottom: 15px; }
.info-label { font-size: 0.7rem; font-weight: 700; color: var(--text-muted); }
.info-value { font-family: monospace; font-size: 0.9rem; font-weight: 700; word-break: break-all; }

button, .btn {
  width: 100%;
  padding: 16px;
  background: var(--text);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  display: block;
}

button:hover, .btn:hover { background: var(--accent); transform: scale(1.02); }
button:active, .btn:active { transform: scale(0.98); }

input[type="file"] { margin: 20px 0; font-size: 0.9rem; width: 100%; }
input[type="password"], textarea {
  width: 100%;
  padding: 12px;
  background: #F0F0F0;
  border: 2px solid var(--border);
  border-radius: 10px;
  font-family: monospace;
  margin-bottom: 15px;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.footer { margin-top: 30px; font-size: 0.7rem; font-weight: 700; opacity: 0.5; letter-spacing: 1px; }
`

func baseLayout(title, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Godrop - %s</title>
    <style>%s</style>
</head>
<body>
    <div class="retro-container">
        <h1>GODROP</h1>
        %s
        <div class="footer">POWERED BY GODROP LOCAL</div>
    </div>
</body>
</html>`, title, sharedCSS, content)
}

// GetSendTemplate renders the Send landing page
func GetSendTemplate(fileName, fileSize string, hasPassword bool) string {
	content := fmt.Sprintf(`
		<p>Sharing a file with you at light speed.</p>
		<div class="info-card">
			<div class="info-item">
				<div class="info-label">FILE</div>
				<div class="info-value">%s</div>
			</div>
			<div class="info-item">
				<div class="info-label">SIZE</div>
				<div class="info-value">%s</div>
			</div>
		</div>
	`, fileName, fileSize)

	if hasPassword {
		content += `
			<input type="password" id="pass" placeholder="ENTER PASSWORD">
			<div id="msg" style="color:var(--accent-bright); font-size:0.7rem; margin-bottom:10px; font-weight:700;"></div>
			<button onclick="verify()">UNLOCK & DOWNLOAD</button>
			<script>
				async function verify() {
					const c = document.getElementById('pass').value;
					const r = await fetch('/api/verify', {method:'POST', body:JSON.stringify({Code:c})});
					const j = await r.json();
					if(j.success) window.location.href='/download';
					else document.getElementById('msg').innerText = "ACCESS DENIED";
				}
			</script>
		`
	} else {
		content += `<a href="/download" class="btn">DOWNLOAD NOW</a>`
	}

	return baseLayout("Download", content)
}

// GetReceiveTemplate renders the Receive landing page
func GetReceiveTemplate() string {
	content := `
		<p>Ready to receive. Drop your file into the box below to send it to the PC.</p>
		<form action="/upload" method="post" enctype="multipart/form-data">
			<div class="info-card" style="text-align:center;">
				<input type="file" name="file" required id="file-input">
				<label for="file-input" style="display:none;">Choose File</label>
			</div>
			<button type="submit">SEND TO PC</button>
		</form>
	`
	return baseLayout("Receive", content)
}

// GetClipboardTemplate renders the Clipboard landing page
func GetClipboardTemplate(currentText string) string {
	// Replacing direct interpolation with a safe approach for scripts
	safeText := strings.ReplaceAll(currentText, "`", "\\`")

	content := fmt.Sprintf(`
		<p>Real-time clipboard synchronization.</p>
		<form action="/clipboard" method="POST" style="display:flex; flex-direction:column;">
			<textarea name="text" id="txt" rows="10" placeholder="Type here...">%s</textarea>
			<div style="display:grid; grid-template-columns:1fr 1fr; gap:10px; margin-top:10px;">
				<button type="button" id="btn-copy" onclick="copyToPhone()" style="background:var(--accent);">COPY TO PHONE</button>
				<button type="submit">SEND TO PC</button>
			</div>
		</form>
		<script>
			const txt = document.getElementById('txt');
			const initialText = `+"`"+`%s`+"`"+`;
			setInterval(async () => {
				if (document.activeElement === txt) return;
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
				const btn = document.getElementById('btn-copy');
				const old = btn.innerText;
				btn.innerText = "COPIED!";
				setTimeout(() => btn.innerText = old, 1500);
			}
		</script>
	`, currentText, safeText)
	return baseLayout("Clipboard", content)
}
