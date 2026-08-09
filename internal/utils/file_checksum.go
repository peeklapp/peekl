package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func GetMd5CheckumForFile(filePath string, root *os.Root) (string, error) {
	var file *os.File
	var err error

	if root != nil {
		file, err = root.Open(filePath)
	} else {
		file, err = os.Open(filePath)
	}
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	file.Seek(0, 0)
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
