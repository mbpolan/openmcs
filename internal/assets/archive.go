package assets

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type archiveFile struct {
	compressedSize   int
	decompressedSize int
	compressed       bool
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
		data, err = decompress(data)
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
	cur, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	offset := int(int(cur) + int(numFiles)*10)
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
			compressed:       compressedSize < decompressedSize,
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
	hash := int32(0)

	// compute a hash of the filename
	str := strings.ToUpper(name)
	for _, ch := range str {
		hash = (hash*0x3D + ch) - 0x20
	}

	// find the data for this file in the archive
	f, ok := a.files[int(hash)]
	if !ok {
		return nil, fmt.Errorf("no such file %s with hash %d in archive", name, hash)
	}

	// decompress the file if necessary
	if f.compressed {
		data := make([]byte, f.compressedSize)
		start := 0

		for i := f.offset; i < f.offset+f.compressedSize; i++ {
			data[start] = a.data[i]
			start++
		}

		return decompress(data)
	}

	// since it's not compressed we can just copy the slice starting from the offset
	data := make([]byte, f.decompressedSize)
	for i := f.offset; i < f.decompressedSize; i++ {
		data[i] = a.data[i]
	}

	return data, nil
}

func decompress(data []byte) ([]byte, error) {
	// prepend a bzip2 header to the byte slice
	compressedData := append([]byte{0x42, 0x5A, 0x68, 0x31}, data...)
	br := bzip2.NewReader(bytes.NewReader(compressedData))

	data, err := io.ReadAll(br)
	if err != nil {
		return nil, err
	}

	return data, nil
}
