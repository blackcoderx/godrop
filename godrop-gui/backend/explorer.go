package backend

import (
	"os"
	"path/filepath"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetHomeDir returns the user's home directory
func (c *Core) GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// ReadDir lists files in a directory
func (c *Core) ReadDir(path string) ([]FileEntry, error) {
	if (path == "" || path == "root") && runtime.GOOS == "windows" {
		var drives []FileEntry
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drivePath := string(drive) + ":\\"
			_, err := os.Stat(drivePath)
			if err == nil {
				drives = append(drives, FileEntry{
					Name:  drivePath,
					Path:  drivePath,
					IsDir: true,
					Type:  "folder",
					Size:  "",
				})
			}
		}
		return drives, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, e := range entries {
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}

		info, err := e.Info()
		size := ""
		if err == nil {
			size = FormatSize(info.Size())
		}

		entry := FileEntry{
			Name:  e.Name(),
			Path:  filepath.Join(path, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
			Type:  "file",
		}
		if e.IsDir() {
			entry.Type = "folder"
			entry.Size = ""
		}
		files = append(files, entry)
	}
	return files, nil
}

// GetDefaultSaveDir returns the user's Downloads directory
func (c *Core) GetDefaultSaveDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Downloads")
}

// SelectDirectory opens a dialog to select a directory
func (c *Core) SelectDirectory() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(c.Ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Save Location",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}
