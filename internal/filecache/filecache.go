package filecache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/sync/singleflight"
)

type FileInfo struct {
	Hash    string
	ModTime int64
	Size    int64
}

type FileCache struct {
	mu      sync.RWMutex
	entries map[string]FileInfo
	group   singleflight.Group
}

func (f *FileCache) GetInfo(root *os.Root, path string, role string) (FileInfo, error) {
	// Get current file information
	stat, err := root.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	f.mu.RLock()
	cached, ok := f.entries[fmt.Sprintf("%s-%s", role, path)]
	f.mu.RUnlock()

	if ok && cached.ModTime == stat.ModTime().UnixNano() && cached.Size == stat.Size() {
		return cached, nil
	}

	v, err, _ := f.group.Do(fmt.Sprintf("%s-%s", role, path), func() (any, error) {
		return f.computeAndStore(root, path, role)
	})
	if err != nil {
		return FileInfo{}, err
	}

	return v.(FileInfo), nil
}

func (f *FileCache) computeAndStore(root *os.Root, path string, role string) (FileInfo, error) {
	stat, err := root.Stat(path)
	if err != nil {
		f.mu.RLock()
		cached, ok := f.entries[fmt.Sprintf("%s-%s", role, path)]
		f.mu.RUnlock()
		if ok && cached.ModTime == stat.ModTime().UnixNano() && cached.Size == stat.Size() {
			return cached, nil
		}
	}

	file, err := root.Open(path)
	if err != nil {
		return FileInfo{}, err
	}
	defer file.Close()

	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return FileInfo{}, err
	}

	info := FileInfo{
		Hash:    hex.EncodeToString(h.Sum(nil)),
		ModTime: stat.ModTime().UnixNano(),
		Size:    stat.Size(),
	}

	f.mu.Lock()
	f.entries[fmt.Sprintf("%s-%s", role, path)] = info
	f.mu.Unlock()

	return info, nil
}

func New() *FileCache {
	return &FileCache{
		entries: make(map[string]FileInfo),
	}
}
