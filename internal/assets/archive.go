package assets

// Archive is a collection of related assets.
type Archive struct {
	data []byte
}

// NewArchive returns a new handle for an archive with data.
func NewArchive(data []byte) *Archive {
	return &Archive{data: data}
}
