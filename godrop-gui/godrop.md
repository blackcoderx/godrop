# Godrop GUI - Project Documentation

Godrop is a high-performance local file transfer and clipboard synchronization tool built with **Wails v2**, **Go**, and **React**.

## Architecture Overview

Godrop follows a classic Wails architecture where the Go backend handles system-level operations and the HTTP server, while the React frontend provides a retro-themed user interface.

### File Structure (godrop-gui)

-   `main.go`: Entry point. Configures Wails application options (window size, title, bindings).
-   `app.go`: The core backend logic. Contains the `App` struct and its bound methods.
    -   **File Navigation**: Methods for reading directories and handling Windows drives.
    -   **Server Management**: Handles starting/stopping Send, Receive, and Clipboard servers.
    -   **Smart Port Management**: Automatically finds an available port if the requested one is occupied.
    -   **Progress Tracking**: Uses a custom `ProgressTracker` to emit real-time transfer stats via Wails events.
-   `frontend/src/App.jsx`: The main UI component.
    -   Manages application state (files, modes, logs, server status).
    -   Implements a 3-way mode switch: **SEND**, **RECEIVE**, and **CLIPBOARD**.
    -   Listens for backend events (progress, file receipt, server status).
-   `frontend/src/App.css`: Implementation of the **Solarized Light** retro theme.

## Key Features

### 1. Send Mode 🚀
-   **Multiple File Selection**: Select one or many files. If multiple are chosen, the backend automatically zips them into a temporary archive.
-   **Download Limits**: Configure how many times a file can be downloaded before the server shuts down.
-   **Time Limits**: Set an expiry timer (in minutes) for the transfer link.
-   **Password Protection**: Optional password for secure downloads.
-   **QR Integration**: Generates a QR code for easy mobile access.

### 2. Receive Mode 📥
-   **Dropzone**: Hosts a web-based upload form.
-   **Direct Streaming**: Files are streamed directly to the disk to avoid memory overhead for large transfers.
-   **Configurable Save Location**: Default to `Downloads`, but user-changeable via native dialog.

### 3. Shared Clipboard 📋
-   **Bi-directional Sync**: Real-time text synchronization between the desktop and mobile devices.
-   **Live Polling**: The mobile interface polls the server to keep the text area updated, while the desktop UI polls the system clipboard.

### 4. Smart Port Management ⚙️
-   The app tries to start on the user's preferred port (default 8080).
-   If the port is in use, it automatically increments and tests up to 100 ports until it finds an available one.
-   The UI automatically updates to reflect the actual port used.

### 5. Real-Time Progress Tracking 📊
-   **Visual Feedback**: A retro-style progress bar appears during active transfers.
-   **Throttled Updates**: Progress events are emitted to the frontend every 100ms to ensure smooth UI performance without overloading the event bus.

## How it works (The Data Flow)

1.  **Preparation**: The user selects a mode and configures settings in the React frontend.
2.  **Startup**: Clicking "START" calls a Go method (`StartServer`, etc.).
3.  **Collision Handling**: Go checks for port availability and starts an `http.Server` in a background goroutine.
4.  **Interaction**: Connected devices access the server via the IP and Port displayed.
5.  **Tracking**: During transfer, `io.Copy` is wrapped by `ProgressTracker`, which sends percentage and byte counts back to React via `EventsEmit`.
6.  **Cleanup**: When the server stops (either manually or via limits), temporary files (like zip archives) are deleted.
