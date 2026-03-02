//go:build windows

package clip

import (
	"os/exec"
	"strings"
)

func Supported() bool { return true }

func CopyText(s string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard")
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}
