package main

import (
	"context"
	"godrop-gui/backend"
	"godrop-gui/backend/server"
)

// App struct acts as a facade between Wails and the backend logic
type App struct {
	core *backend.Core
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		core: backend.NewCore(),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.core.Ctx = ctx
	go a.core.MonitorClipboard()
}

// --- FILE SYSTEM ---

func (a *App) GetHomeDir() string {
	return a.core.GetHomeDir()
}

func (a *App) ReadDir(path string) ([]backend.FileEntry, error) {
	return a.core.ReadDir(path)
}

func (a *App) GetDefaultSaveDir() string {
	return a.core.GetDefaultSaveDir()
}

func (a *App) SelectDirectory() (string, error) {
	return a.core.SelectDirectory()
}

// --- CLIPBOARD ---

func (a *App) GetSystemClipboard() string {
	return a.core.GetSystemClipboard()
}

func (a *App) SetSystemClipboard(text string) {
	a.core.SetSystemClipboard(text)
}

func (a *App) GetHistory() []string {
	return a.core.GetHistory()
}

// --- SERVERS ---

func (a *App) StartServer(port string, password string, files []string, limit int, timeout int) (server.ServerResponse, error) {
	a.core.ServerMutex.Lock()
	defer a.core.ServerMutex.Unlock()

	if a.core.Server != nil {
		server.Stop(a.core)
	}

	return server.StartSend(a.core, port, password, files, limit, timeout)
}

func (a *App) StartReceiveServer(port string, saveDir string) (server.ServerResponse, error) {
	a.core.ServerMutex.Lock()
	defer a.core.ServerMutex.Unlock()

	if a.core.Server != nil {
		server.Stop(a.core)
	}

	return server.StartReceive(a.core, port, saveDir)
}

func (a *App) StartClipboardServer(port string) (server.ServerResponse, error) {
	a.core.ServerMutex.Lock()
	defer a.core.ServerMutex.Unlock()

	if a.core.Server != nil {
		server.Stop(a.core)
	}

	return server.StartClipboard(a.core, port)
}

func (a *App) StopServer() {
	a.core.ServerMutex.Lock()
	defer a.core.ServerMutex.Unlock()
	server.Stop(a.core)
}
