//go:build !darwin && !windows

package clip

import "errors"

func Supported() bool { return false }

func CopyText(_ string) error {
	return errors.New("clipboard not supported on this platform")
}
