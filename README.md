# GODROP // SYSTEM_V1.0

![GoDrop Retro](https://img.shields.io/badge/GODROP-V1.0-00ff41?style=for-the-badge&logo=go&logoColor=white) 
![License](https://img.shields.io/badge/license-MIT-00ff41?style=for-the-badge)

**GoDrop** is a minimal, vintage-inspired file-sharing tool that converts your local machine into a temporary "SaaS" share point. Share files instantly across devices on the same WiFi network—your computer runs the server, other devices (phones, tablets, laptops) download via a beautiful retro web interface.

---

## 📟 Features

- **Retro Aesthetic**: A classic CRT/Terminal landing page with real-time stats
- **Multiple Files**: Automatically zips multiple files for the recipient
- **Zero Configuration**: Just run it and scan the QR code
- **Security Codes**: Protect your shares with a PIN/Access Code
- **Download Limits**: Automatically shut down the server after N downloads
- **Auto-Expiry**: Set a time limit for how long files are available
- **Cross-Platform**: Works on Windows, Linux, and macOS

---

## 🚦 How It Works

1. **Run the command** with the files you want to share
2. **Scan the QR code** that appears in your terminal
3. **Beautiful UI loads** on your phone/tablet with file info
4. **Download the files** from the retro interface
5. **Server auto-terminates** when download limit is reached

**Important**: Both devices must be on the **same WiFi network**.

---

## 🚀 Installation

### Prerequisites
- **[Go 1.20 or higher](https://go.dev/doc/install)** (for building from source)
- Both devices on the **same WiFi network**
- Firewall configured to allow the application (see [Troubleshooting](#troubleshooting))

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/blackcoderx/godrop.git
cd godrop

# Download dependencies
go mod download

# Build the executable
go build -o godrop
```

### Option 2: Download Pre-Built Binary
*(Coming soon - check [Releases](https://github.com/blackcoderx/godrop/releases))*

---

## 📂 Important: Project Structure

When running GoDrop, the executable **must** be in the same directory as the `web` folder:

```
godrop/
├── godrop.exe (or godrop on Linux/Mac)
└── web/
    ├── index.html
    ├── script.js
    └── style.css
```

**If deploying to another computer**, copy both:
- The executable (`godrop` or `godrop.exe`)
- The entire `web/` folder

---

## 🛠 Usage

### Basic Usage
Share a single file:
```bash
./godrop myfile.pdf
```

Share multiple files (auto-zipped):
```bash
./godrop file1.pdf file2.png file3.txt
```

### Advanced Usage
Professional "SaaS" mode with security and limits:
```bash
./godrop -limit 3 -code 1234 -timeout 30m file1.pdf file2.png
```

### Command Flags

| Flag | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| `-limit` | Max downloads before shutdown | `1` | `-limit 5` |
| `-code` | Security PIN for access | *(none)* | `-code "PASS123"` |
| `-timeout` | Time limit (m=minutes, h=hours) | *(none)* | `-timeout 1h` |
| `-port` | Custom server port | `8080` | `-port 9090` |

### Example Output
```
----------------------------------------
█████████████████████████████████████
█████████████████████████████████████
███   QR CODE APPEARS HERE        ███
█████████████████████████████████████
█████████████████████████████████████

Hosting: godrop-archive.zip
Downloads Allowed: 3
Security Code REQUIRED: 1234
Expiry Time: 18:35:25
Share Link: http://192.168.1.15:8080
----------------------------------------
GODROP Server Live on :8080
```

---

## 💻 Multi-Platform Build Guide

Compile GoDrop for any platform from your current machine.

### Windows (64-bit)
```bash
GOOS=windows GOARCH=amd64 go build -o godrop.exe .
```

### Linux (64-bit)
```bash
GOOS=linux GOARCH=amd64 go build -o godrop .
```

### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o godrop .
```

### macOS (Apple Silicon M1/M2/M3)
```bash
GOOS=darwin GOARCH=arm64 go build -o godrop .
```

**Remember**: Always include the `web/` folder when deploying to another computer!

---

## ✅ Verify Installation

Test that everything works:

```bash
# Create a test file
echo "Hello from GoDrop!" > test.txt

# Run GoDrop
./godrop test.txt

# Expected: QR code appears, server starts on port 8080
# Open http://localhost:8080 in your browser to see the retro UI
```

If you see the retro interface, installation is successful! ✅

---

## 🎨 Retro UI

Recipients see a high-fidelity retro terminal interface featuring:
- **Live Stats**: Real-time download counter and file size display
- **Integrated Auth**: Security code verification (if enabled)
- **Countdown Timer**: Shows time remaining before expiry
- **No Refresh Needed**: All updates via WebSocket-like polling
- **One-Click Download**: Big, obvious download button

---

## 🔧 Troubleshooting

### "Connection Refused" or QR Code Won't Load

**Problem**: Firewall is blocking the application.

**Solutions**:
- **Windows**: Allow `godrop.exe` through Windows Defender Firewall
  - Settings → Privacy & Security → Windows Security → Firewall → Allow an app
- **macOS**: System Preferences → Security & Privacy → Firewall → Allow godrop
- **Linux**: Check `ufw` or `iptables` rules

### QR Code Scans But Page Won't Load

**Problem**: Devices are not on the same network.

**Solutions**:
- Ensure both devices are connected to the **same WiFi network**
- Disable VPN on either device
- Try accessing the IP address manually instead of scanning QR code

### "Port Already in Use"

**Problem**: Another application is using port 8080.

**Solution**: Use a different port:
```bash
./godrop -port 9090 myfile.pdf
```

### Web Folder Not Found

**Problem**: `web/` directory is missing or in wrong location.

**Solution**: 
- Ensure `web/` folder is in the same directory as the executable
- If you moved the executable, copy the `web/` folder too

### "Cannot Find Module github.com/skip2/go-qrcode"

**Problem**: Dependencies not downloaded.

**Solution**:
```bash
go mod download
go build -o godrop
```

---

## 🔒 Security Notes

- GoDrop is designed for **trusted local networks** (home/office WiFi)
- Security codes provide basic protection but are **not encrypted**
- Files are transferred over **HTTP (not HTTPS)** on your local network
- Server auto-terminates after download limit or timeout
- Do not expose the server to the public internet

---

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Report bugs via [Issues](https://github.com/blackcoderx/godrop/issues)
- Submit pull requests
- Suggest new features

---

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 🙏 Acknowledgments

- QR Code generation: [go-qrcode](https://github.com/skip2/go-qrcode)
- Retro terminal aesthetic inspired by classic computing
