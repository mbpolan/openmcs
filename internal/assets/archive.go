package assets

import (
	"bytes"
	"compress/bzip2"
	"github.com/pkg/errors"
	"io"
)

type archiveFile struct {
	compressedSize   int
	decompressedSize int
	offset           int
}

// Archive is a collection of related assets.
type Archive struct {
	data  []byte
	files map[int]archiveFile
}

// NewArchive returns a new handle for an archive with data.
func NewArchive(data []byte) (*Archive, error) {
	// read archive properties from the raw data
	r := NewDataReader(data)
	compressedSize, err := r.Uint24()
	if err != nil {
		return nil, err
	}

	decompressedSize, err := r.Uint24()
	if err != nil {
		return nil, err
	}

	// is this archive compressed? if so, we need to decompress the data first
	if compressedSize != decompressedSize {
		compressedData := append([]byte{'B', 'Z', 'h', '1'}, data...)
		br := bzip2.NewReader(bytes.NewReader(compressedData))

		data, err = io.ReadAll(br)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decompress file")
		}

		// start reading from the beginning of the decompressed data
		r = NewDataReader(data)
	}

	// read the number of files in this archive
	numFiles, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// start an initial offset from here
	offset := int(2 + numFiles*10)
	files := map[int]archiveFile{}

	// read information for all files in the archive
	for i := 0; i < int(numFiles); i++ {
		hash, err := r.Uint32()
		if err != nil {
			return nil, err
		}

		decompressedSize, err := r.Uint24()
		if err != nil {
			return nil, err
		}

		compressedSize, err := r.Uint24()
		if err != nil {
			return nil, err
		}

		files[int(hash)] = archiveFile{
			compressedSize:   int(compressedSize),
			decompressedSize: int(decompressedSize),
			offset:           offset,
		}

		offset += int(compressedSize)
	}

	return &Archive{
		data:  data,
		files: files,
	}, nil
}

// File returns the data associated with a file in the archive.
func (a *Archive) File(name string) ([]byte, error) {
	return nil, nil
}
