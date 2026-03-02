//go:build darwin

package clip

import (
	"os/exec"
	"strings"
)

func Supported() bool { return true }

func CopyText(s string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}
