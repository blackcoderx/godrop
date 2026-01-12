package backend

import (
	"github.com/atotto/clipboard"
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
}
