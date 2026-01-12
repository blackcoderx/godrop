package backend

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

// GetOutboundIP returns the preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// FindAvailablePort tries to find an available port starting from startPort
func FindAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("could not find an available port after 100 attempts")
}

// CreateZipArchive zips multiple files or a directory into a temporary archive
func (c *Core) CreateZipArchive(files []string) (string, string, error) {
	tmpZip, err := os.CreateTemp("", "godrop-*.zip")
	if err != nil {
		return "", "", err
	}
	zw := zip.NewWriter(tmpZip)
	for _, f := range files {
		if err := AddFileToZip(zw, f, ""); err != nil {
			log.Printf("Failed to add %s: %v", f, err)
		}
	}
	zw.Close()
	tmpZip.Close()

	c.IsTempArchive = true
	c.ArchivePath = tmpZip.Name()

	name := "godrop-archive.zip"
	if len(files) == 1 {
		name = filepath.Base(files[0]) + ".zip"
	}

	return tmpZip.Name(), name, nil
}

// AddFileToZip is a recursive helper to add files or directories to a zip writer
func AddFileToZip(zw *zip.Writer, fullPath string, baseInZip string) error {
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	if baseInZip == "" {
		header.Name = filepath.Base(fullPath)
	} else {
		header.Name = filepath.Join(baseInZip, filepath.Base(fullPath))
	}

	if info.IsDir() {
		header.Name += "/"
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		files, err := os.ReadDir(fullPath)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := AddFileToZip(zw, filepath.Join(fullPath, f.Name()), header.Name); err != nil {
				return err
			}
		}
		_ = writer
		return nil
	}

	header.Method = zip.Deflate
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

// FormatSize formats bytes into a human-readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ToBase64 helper
func ToBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
