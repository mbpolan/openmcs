package asset

import (
	"github.com/mbpolan/openmcs/internal/model"
	"path"
)

const (
	cacheMain int = 0
	cacheMap      = 4
)

const (
	archiveConfig   int = 2
	archiveVersions     = 5
)

// Manager handles loading and managing game assets.
type Manager struct {
	baseDir  string
	archives map[int]*Archive
	caches   map[int]*CacheFile
}

// NewManager returns a new asset Manager instance with game assets located at the given baseDir. You should call
// the Close() method once all assets have been loaded to free up resources.
func NewManager(baseDir string) *Manager {
	// get a handle to the cache files
	dir := path.Join(baseDir, "data")

	return &Manager{
		baseDir:  dir,
		archives: map[int]*Archive{},
		caches:   map[int]*CacheFile{},
	}
}

// Close releases all resources related to located assets.
func (m *Manager) Close() {
	m.archives = map[int]*Archive{}
	m.caches = map[int]*CacheFile{}
}

// Map returns the world map extracted from game assets.
func (m *Manager) Map() (*model.Map, error) {
	versions, err := m.archive(cacheMain, archiveVersions)
	if err != nil {
		return nil, err
	}

	mapLoader := NewMapLoader(versions, m.cache(cacheMap))
	worldMap, err := mapLoader.Load()
	if err != nil {
		return nil, err
	}

	return worldMap, nil
}

// Items returns a slice of model.Item data extracted from game assets.
func (m *Manager) Items() ([]*model.Item, error) {
	archive, err := m.archive(cacheMain, archiveConfig)
	if err != nil {
		return nil, err
	}

	itemLoader := NewItemLoader(archive)
	return itemLoader.Load()
}

// WorldObjects returns a slice of model.WorldObject data extracted from game assets.
func (m *Manager) WorldObjects() ([]*model.WorldObject, error) {
	archive, err := m.archive(cacheMain, archiveConfig)
	if err != nil {
		return nil, err
	}

	// load world objects
	woLoader := NewWorldObjectLoader(archive)
	return woLoader.Load()
}

// cache returns a CacheFile for the given index ID.
func (m *Manager) cache(id int) *CacheFile {
	if cache, ok := m.caches[id]; ok {
		return cache
	}

	cache := NewCacheFile(m.baseDir, id)
	m.caches[id] = cache
	return cache
}

// archive returns an Archive for the given archive ID.
func (m *Manager) archive(cacheID int, id int) (*Archive, error) {
	cache := m.cache(cacheID)

	if archive, ok := m.archives[id]; ok {
		return archive, nil
	}

	archive, err := cache.Archive(id)
	if err != nil {
		return nil, err
	}

	m.archives[id] = archive
	return archive, nil
}
