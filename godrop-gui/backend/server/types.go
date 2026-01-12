package server

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// ServerResponse returns to the frontend
type ServerResponse struct {
	IP      string `json:"ip"`
	Port    string `json:"port"`
	FullURL string `json:"fullUrl"`
	QRCode  string `json:"qrCode"` // Base64 encoded PNG
}

// Dedicated context for server state management
type ServerState struct {
	Server           *http.Server
	Mutex            sync.Mutex
	DownloadLimit    int
	CurrentDownloads int
	ExpiryTime       time.Time
	Ctx              context.Context
}
