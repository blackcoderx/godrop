package server

import (
	"context"
	"os"

	"godrop-gui/backend"
)

// Stop shuts down the server and cleans up resources
func Stop(core *backend.Core) {
	if core.Server != nil {
		core.Server.Shutdown(context.Background())
		core.Server = nil
	}
	if core.IsTempArchive && core.ArchivePath != "" {
		os.Remove(core.ArchivePath)
		core.IsTempArchive = false
	}
}
