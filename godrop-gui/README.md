# Godrop 🚀

**Godrop** is a high-speed, local-first file sharing and clipboard synchronization tool built with **Wails**, **Go**, and **React**. It features a unique **Vibrant Retro** aesthetic and is designed for maximum security and simplicity within your local network.

![Vibrant Retro Style](https://img.shields.io/badge/Style-Vibrant_Retro-bluebell)
![Local Only](https://img.shields.io/badge/Connection-Local_Only-success)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8)
![Wails](https://img.shields.io/badge/Framework-Wails_v2-red)

## ✨ Features

- **📂 Fast File Sharing**: Direct peer-to-peer file transfer over your local Wi-Fi. No cloud, no middleman.
- **📥 Dropzone (Receive Mode)**: Set up a folder as a dropzone and let other devices upload files directly to your PC.
- **📋 Shared Clipboard**: Sync your clipboard across devices in real-time. Copy on your PC, paste on your phone, and vice-versa.
- **♾️ Infinity Mode**: Keep your server live indefinitely by setting the timeout to `0`.
- **🔐 Secure by Default**: Optional password protection for file transfers.
- **🎨 Vibrant Retro UI**: A minimal, animated interface with a bold retro palette (White Smoke, Blue Bell, Retro Coral).
- **🌐 Unified Web Experience**: Landing pages for external devices perfectly match the desktop application's aesthetic.

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later
- [Node.js](https://nodejs.org/) & npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Installation & Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/blackcoderx/godrop.git
   cd godrop/godrop-gui
   ```

2. **Run in Development Mode**:
   ```bash
   wails dev
   ```
   This will start the Wails development server with hot-reload for the React frontend and Go backend.

3. **Build for Production**:
   ```bash
   wails build
   ```
   The compiled executable will be located in the `build/bin` directory.

## 📁 Project Structure

Godrop follows a clean, modular architecture:

- `backend/`: Core Go logic.
    - `server/`: Specialized HTTP servers for Send, Receive, and Clipboard modes.
    - `explorer.go`: File system navigation.
    - `progress.go`: Real-time transfer tracking.
- `frontend/`: React application.
    - `src/components/`: Modular UI components (Tabs, Config, Overlay).
    - `src/App.css`: The central "Vibrant Retro" design system.
- `app.go`: The Wails bridge/facade connecting the frontend to the modular backend.

## 🛡️ Privacy

Godrop is **local-only**. Your files and clipboard data never leave your local network. There are no cloud servers involved, ensuring your data stays private and transfers at the maximum speed supported by your Wi-Fi.

---

> [!TIP]
> Use Port `1111` for a memorable connection! 🔌✨
