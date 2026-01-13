package backend

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetSystemClipboard returns the current system clipboard text
func (c *Core) GetSystemClipboard() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	return text
}

// SetSystemClipboard sets the system clipboard text
func (c *Core) SetSystemClipboard(text string) {
	clipboard.WriteAll(text)
	c.AddToHistory(text)
}

// AddToHistory adds a new item to the clipboard history
func (c *Core) AddToHistory(text string) {
	if text == "" {
		return
	}
	c.ClipboardMutex.Lock()
	defer c.ClipboardMutex.Unlock()

	// Avoid duplicates at the top
	if len(c.ClipboardHistory) > 0 && c.ClipboardHistory[0] == text {
		return
	}

	// Prepend to history
	c.ClipboardHistory = append([]string{text}, c.ClipboardHistory...)

	// Limit history size (e.g., 50 items)
	if len(c.ClipboardHistory) > 50 {
		c.ClipboardHistory = c.ClipboardHistory[:50]
	}
}

// GetHistory returns the clipboard history
func (c *Core) GetHistory() []string {
	c.ClipboardMutex.Lock()
	defer c.ClipboardMutex.Unlock()
	return c.ClipboardHistory
}

// MonitorClipboard watches the system clipboard for changes
func (c *Core) MonitorClipboard() {
	lastText := c.GetSystemClipboard()
	c.AddToHistory(lastText)

	for {
		time.Sleep(1 * time.Second)
		currentText := c.GetSystemClipboard()
		if currentText != "" && currentText != lastText {
			lastText = currentText
			c.AddToHistory(currentText)
			// Emit event to frontend if Ctx is available
			if c.Ctx != nil {
				runtime.EventsEmit(c.Ctx, "clipboard-changed", currentText)
			}
		}
	}
}
