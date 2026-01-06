# GODROP // SYSTEM_V1.0

![GoDrop Retro](https://img.shields.io/badge/GODROP-V1.0-00ff41?style=for-the-badge&logo=go&logoColor=white) 
![License](https://img.shields.io/badge/license-MIT-00ff41?style=for-the-badge)

**Godrop** is a minimal, vintage-inspired file-sharing tool that converts your local machine into a temporary "SaaS" share point. Built for developers who like the terminal but need a beautiful interface for the recipient.

---

## 📟 Features

- **Retro Aesthetic**: A classic CRT/Terminal landing page with real-time stats.
- **Multiple Files**: Automatically zips multiple files for the recipient.
- **Zero Configuration**: Just run it and scan the QR code.
- **Security Codes**: Protect your shares with a PIN/Access Code.
- **Download Limits**: Automatically shut down the server after $N$ downloads.
- **Auto-Expiry**: Set a time limit for how long the files are available.
- **Real-time SaaS Dashboard**: The recipient sees a live countdown and download tracker.

---

## 🚀 Installation

### Prerequisites
- [Go](https://go.dev/doc/install) (1.20 or higher)

### Build from source
```bash
# Clone the repository
git clone https://github.com/yourusername/godrop.git
cd godrop

# Install dependencies
go mod download

# Build for your platform
go build -o godrop
```

---

## 🛠 Usage

Basic usage:
```bash
./godrop <file_path>
```

Professional "SaaS" mode:
```bash
./godrop -limit 3 -code 1234 -timeout 30m file1.pdf file2.png
```

### Flags
| Flag | Description | Example |
| :--- | :--- | :--- |
| `-limit` | Number of downloads allowed | `-limit 5` |
| `-code` | Security code for access | `-code "PASS123"` |
| `-timeout` | Time limit (m, h) | `-timeout 1h` |
| `-port` | Custom server port | `-port 9090` |

---

## 💻 Multi-Platform Build Guide

You can compile Godrop for any platform from your current machine.

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o godrop.exe .
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o godrop .
```

### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o godrop .
```

### macOS (Apple Silicon/M1/M2)
```bash
GOOS=darwin GOARCH=arm64 go build -o godrop .
```

---

## 🎨 Retro UI

The recipient will be greeted with a high-fidelity retro terminal interface.
- **Live Sync**: No page refresh needed.
- **Integrated Auth**: Access codes verified via internal API.
- **Countdown**: Smooth ticking system-clock.

---

## 📝 License
Distributed under the MIT License. See `LICENSE` for more information.
