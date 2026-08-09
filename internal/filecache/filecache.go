package filecache

import (
	"os"
	"sync"

	"github.com/peeklapp/peekl/internal/utils"
	"golang.org/x/sync/singleflight"
)

type FileInfo struct {
	Hash    string
	ModTime int64
}

type FileCache struct {
	mu      sync.RWMutex
	entries map[string]FileInfo
	group   singleflight.Group
}

func (f *FileCache) GetInfo(root *os.Root, path string) (FileInfo, error) {
	stat, err := root.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	f.mu.RLock()
	cached, ok := f.entries[path]
	f.mu.RUnlock()

	if ok && cached.ModTime == stat.ModTime().UnixNano() {
		return cached, nil
	}

	v, err, _ := f.group.Do(path, func() (any, error) {
		return f.computeAndStore(root, path)
	})
	if err != nil {
		return FileInfo{}, err
	}

	return v.(FileInfo), nil
}

func (f *FileCache) computeAndStore(root *os.Root, path string) (FileInfo, error) {
	stat, err := root.Stat(path)
	if err != nil {
		f.mu.RLock()
		cached, ok := f.entries[path]
		f.mu.RUnlock()
		if ok && cached.ModTime == stat.ModTime().UnixNano() {
			return cached, nil
		}
	}

	checksum, err := utils.GetMd5CheckumForFile(path, root)
	if err != nil {
		return FileInfo{}, err
	}

	info := FileInfo{
		Hash:    checksum,
		ModTime: stat.ModTime().UnixNano(),
	}

	f.mu.Lock()
	f.entries[path] = info
	f.mu.Unlock()

	return info, nil
}

func New() *FileCache {
	return &FileCache{
		entries: make(map[string]FileInfo),
	}
}
