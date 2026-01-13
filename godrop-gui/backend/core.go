package backend

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// FileEntry represents a file in the explorer
type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"img"` // Key 'img' matches current frontend expectation
	IsDir bool   `json:"isDir"`
	Size  string `json:"size"`
	Type  string `json:"type"` // "file" or "folder"
}

// Core holds the application state and logic
type Core struct {
	Ctx              context.Context
	Server           *http.Server
	ServerMutex      sync.Mutex
	IsTempArchive    bool
	ArchivePath      string
	DownloadLimit    int
	CurrentDownloads int
	ExpiryTime       time.Time
	ClipboardHistory []string
	ClipboardMutex   sync.Mutex
}

func NewCore() *Core {
	return &Core{}
}
