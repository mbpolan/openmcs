package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const ReconnectLoginRequestHeader byte = 0x10
const NewLoginRequestHeader byte = 0x12

// LoginRequest is sent by the client when a player attempts to log into the server.
type LoginRequest struct {
	Seeds       []uint32
	Version     uint16
	UID         uint32
	Username    string
	Password    string
	IsLowMemory bool
	CRCs        []uint32
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *LoginRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// skip unknown bytes
	err = r.Skip(2)
	if err != nil {
		return err
	}

	// read client version
	version, err := r.Uint16()
	if err != nil {
		return err
	}

	// read low memory indicator
	lowMemory, err := r.Uint8()
	if err != nil {
		return err
	}

	// read expected cache crcs
	crcs := make([]uint32, 9)
	for i := 0; i < len(crcs); i++ {
		crc, err := r.Uint32()
		if err != nil {
			return err
		}

		crcs[i] = crc
	}

	// read length of remaining buffer
	_, err = r.Uint8()
	if err != nil {
		return err
	}

	// read next segment byte
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	if b != 0x0A {
		return fmt.Errorf("unexpected byte in login request: %2x", b)
	}

	// read four integers containing client seed
	seeds := make([]uint32, 4)
	for i := 0; i < len(seeds); i++ {
		seed, err := r.Uint32()
		if err != nil {
			return err
		}

		seeds[i] = seed
	}

	// client unique identifier
	uid, err := r.Uint32()
	if err != nil {
		return err
	}

	// read the username and password
	username, err := r.String()
	if err != nil {
		return err
	}

	password, err := r.String()
	if err != nil {
		return err
	}

	p.Seeds = seeds
	p.Version = version
	p.UID = uid
	p.Username = username
	p.Password = password
	p.IsLowMemory = lowMemory == 0x01
	p.CRCs = crcs
	return nil
}
