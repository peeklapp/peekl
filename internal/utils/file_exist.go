package utils

import (
	"errors"
	"os"
)

func FileExist(path string, root *os.Root) bool {
	if root != nil {
		if _, err := root.Stat(path); errors.Is(err, os.ErrNotExist) {
			return false
		}
	} else {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}
