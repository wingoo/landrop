package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SaveReader(dir, originalName string, r io.Reader) (string, string, int64, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", 0, err
	}
	safeName := sanitizeFilename(originalName)
	if safeName == "" {
		safeName = "file"
	}
	fullPath, finalName, err := nextAvailablePath(dir, safeName)
	if err != nil {
		return "", "", 0, err
	}

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", "", 0, err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return "", "", n, err
	}
	return fullPath, finalName, n, nil
}

func SaveText(dir, name, text string) (string, string, int64, error) {
	return SaveReader(dir, name, strings.NewReader(text))
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.Trim(name, ". ")
	return name
}

func nextAvailablePath(dir, name string) (string, string, error) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	if base == "" {
		base = "file"
	}
	candidate := name
	for i := 0; ; i++ {
		if i > 0 {
			candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		full := filepath.Join(dir, candidate)
		_, err := os.Stat(full)
		if os.IsNotExist(err) {
			return full, candidate, nil
		}
		if err != nil {
			return "", "", err
		}
	}
}
