package backend

import (
	"context"
	"io"
	"net/http"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProgressTracker tracks io progress and emits Wails events
type ProgressTracker struct {
	Total      int64
	Current    int64
	LastEmit   time.Time
	EventName  string
	Ctx        context.Context
	Writer     io.Writer
	Reader     io.Reader
	IsFinished bool
}

func (pt *ProgressTracker) Write(p []byte) (int, error) {
	n, err := pt.Writer.Write(p)
	pt.Current += int64(n)
	pt.EmitProgress()
	return n, err
}

func (pt *ProgressTracker) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.Current += int64(n)
	pt.EmitProgress()
	return n, err
}

func (pt *ProgressTracker) EmitProgress() {
	if pt.IsFinished {
		return
	}

	percent := int(float64(pt.Current) / float64(pt.Total) * 100)
	if percent > 100 {
		percent = 100
	}

	if percent == 100 || time.Since(pt.LastEmit) > 100*time.Millisecond {
		wailsRuntime.EventsEmit(pt.Ctx, pt.EventName, map[string]interface{}{
			"percent":     percent,
			"transferred": pt.Current,
			"total":       pt.Total,
		})
		pt.LastEmit = time.Now()
	}
}

// ProgressResponseWriter wraps http.ResponseWriter to track bytes written
type ProgressResponseWriter struct {
	http.ResponseWriter
	Tracker *ProgressTracker
}

func (pw *ProgressResponseWriter) Write(p []byte) (int, error) {
	n, err := pw.ResponseWriter.Write(p)
	pw.Tracker.Current += int64(n)
	pw.Tracker.EmitProgress()
	return n, err
}

func (pw *ProgressResponseWriter) Flush() {
	if f, ok := pw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
