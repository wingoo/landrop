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
	ext := filepath.Ext(safeName)
	base := strings.TrimSuffix(safeName, ext)
	if base == "" {
		base = "file"
	}

	for i := 0; ; i++ {
		finalName := safeName
		if i > 0 {
			finalName = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		fullPath := filepath.Join(dir, finalName)
		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", 0, err
		}
		n, copyErr := io.Copy(f, r)
		closeErr := f.Close()
		if copyErr != nil {
			return "", "", n, copyErr
		}
		if closeErr != nil {
			return "", "", n, closeErr
		}
		return fullPath, finalName, n, nil
	}
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
