package assets

import (
	"fmt"
	"io"
	"os"
	"path"
)

const filePrefix = "main_file_cache"
const sectorSize = 520
const sectorHeaderSize = 8

// CacheFile is a handle to the game server cache data.
type CacheFile struct {
	path    string
	index   int
	storeID int
}

// NewCacheFile returns a handle to section of the game server cache data specified by an index.
func NewCacheFile(path string, index int) *CacheFile {
	return &CacheFile{
		path:    path,
		index:   index,
		storeID: index + 1,
	}
}

// Archive returns a handle to an archive located within the game server cache.
func (c *CacheFile) Archive(index int) (*Archive, error) {
	// open the data file
	dataFile, err := os.Open(path.Join(c.path, fmt.Sprintf("%s.dat", filePrefix)))
	if err != nil {
		return nil, err
	}

	dataFileInfo, err := dataFile.Stat()
	if err != nil {
		return nil, err
	}

	// open the index file
	idxFile, err := os.Open(path.Join(c.path, fmt.Sprintf("%s.idx%d", filePrefix, c.index)))
	if err != nil {
		return nil, err
	}

	// seek to the position of the requested archive in the cache index file
	_, err = idxFile.Seek(int64(index*6), 0)
	if err != nil {
		return nil, err
	}

	// read six bytes from the table of contents
	buf := make([]byte, 6)
	_, err = io.ReadFull(idxFile, buf)
	if err != nil {
		return nil, err
	}

	// first three bytes are the archive size, and the following three are the sector where the archive data is
	// located
	size := int((int32(buf[0]) << 16) | (int32(buf[1]) << 8) | int32(buf[2]))
	sector := int((int32(buf[3]) << 16) | (int32(buf[4]) << 8) | int32(buf[5]))

	// validate we didn't read something crazy
	if size < 0 || size > 500000 {
		return nil, fmt.Errorf("invalid archive size found: %d", size)
	}

	if sector <= 0 || int64(sector) > dataFileInfo.Size()/sectorSize {
		return nil, fmt.Errorf("invalid archive sector found: %d", sector)
	}

	data := make([]byte, size)
	read := 0

	for part := 0; read < size; part++ {
		// seek to the current sector location in the data file
		_, err = dataFile.Seek(int64(sector*sectorSize), io.SeekStart)
		if err != nil {
			return nil, err
		}

		// read up to 512 bytes
		remaining := int(size - read)
		if remaining > sectorSize-sectorHeaderSize {
			remaining = sectorSize - sectorHeaderSize
		}

		// TODO: can we reduce the amount of allocations?
		// read the sector from the sector including the header
		chunk := make([]byte, remaining+sectorHeaderSize)
		_, err = io.ReadFull(dataFile, chunk)
		if err != nil {
			return nil, err
		}

		// compute indexes, sector and store identifiers from the header
		r := NewDataReader(chunk)
		realIdx, err := r.Uint16()
		if err != nil {
			return nil, err
		}

		realPart, err := r.Uint16()
		if err != nil {
			return nil, err
		}

		nextSector, err := r.Uint24()
		if err != nil {
			return nil, err
		}

		realStoreID, err := r.Byte()
		if err != nil {
			return nil, err
		}

		// validate the header
		if int(realIdx) != index || int(realPart) != part || int(realStoreID) != c.storeID {
			return nil, fmt.Errorf("invalid chunk found: %d != %d, %d != %d, %d != %d", realIdx, index, realPart, part, nextSector, c.storeID)
		} else if nextSector < 0 || int64(nextSector) > dataFileInfo.Size()/sectorSize {
			return nil, fmt.Errorf("invalid sector found: %d", nextSector)
		}

		for i := 0; i < remaining; i++ {
			data[read+i] = chunk[i+sectorHeaderSize]
		}

		read += remaining
		sector = int(nextSector)
	}

	return NewArchive(data)
}
