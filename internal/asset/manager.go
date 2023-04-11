package asset

import (
	"github.com/mbpolan/openmcs/internal/model"
	"path"
)

const (
	archiveConfig int = 2
)

// Manager handles loading and managing game assets.
type Manager struct {
	cache    *CacheFile
	archives map[int]*Archive
}

// NewManager returns a new asset Manager instance with game assets located at the given baseDir. You should call
// the Close() method once all assets have been loaded to free up resources.
func NewManager(baseDir string) *Manager {
	// get a handle to the main cache file
	cache := NewCacheFile(path.Join(baseDir, "data"), 0)

	return &Manager{
		archives: map[int]*Archive{},
		cache:    cache,
	}
}

// Close releases all resources related to located assets.
func (m *Manager) Close() {
	m.archives = map[int]*Archive{}
	m.cache = nil
}

// Items returns a slice of model.Item data extracted from game assets.
func (m *Manager) Items() ([]*model.Item, error) {
	archive, err := m.archive(archiveConfig)
	if err != nil {
		return nil, err
	}

	itemLoader := NewItemLoader(archive)
	return itemLoader.Load()
}

// WorldObjects returns a slice of model.WorldObject data extracted from game assets.
func (m *Manager) WorldObjects() ([]*model.WorldObject, error) {
	archive, err := m.archive(archiveConfig)
	if err != nil {
		return nil, err
	}

	// load world objects
	woLoader := NewWorldObjectLoader(archive)
	return woLoader.Load()
}

// archive returns an Archive for the given archive ID.
func (m *Manager) archive(id int) (*Archive, error) {
	if archive, ok := m.archives[id]; ok {
		return archive, nil
	}

	archive, err := m.cache.Archive(archiveConfig)
	if err != nil {
		return nil, err
	}

	m.archives[id] = archive
	return archive, nil
}
