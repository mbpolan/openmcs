package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const ReconnectLoginRequestHeader byte = 0x10
const NewLoginRequestHeader byte = 0x12

type LoginRequest struct {
	Seeds       []uint32
	Version     uint16
	UID         uint32
	Username    string
	Password    string
	IsLowMemory bool
	CRCs        []uint32
}

func ReadLoginRequest(r *network.ProtocolReader) (*LoginRequest, error) {
	// skip unknown bytes
	err := r.Skip(2)
	if err != nil {
		return nil, err
	}

	// read client version
	version, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// read low memory indicator
	lowMemory, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	// read expected cache crcs
	crcs := make([]uint32, 9)
	for i := 0; i < len(crcs); i++ {
		crc, err := r.Uint32()
		if err != nil {
			return nil, err
		}

		crcs[i] = crc
	}

	// read length of remaining buffer
	_, err = r.Uint8()
	if err != nil {
		return nil, err
	}

	// read next segment byte
	b, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	if b != 0x0A {
		return nil, fmt.Errorf("unexpected byte in login request: %2x", b)
	}

	// read four integers containing client seed
	seeds := make([]uint32, 4)
	for i := 0; i < len(seeds); i++ {
		seed, err := r.Uint32()
		if err != nil {
			return nil, err
		}

		seeds[i] = seed
	}

	// client unique identifier
	uid, err := r.Uint32()
	if err != nil {
		return nil, err
	}

	// read the username and password
	username, err := r.String()
	if err != nil {
		return nil, err
	}

	password, err := r.String()
	if err != nil {
		return nil, err
	}

	return &LoginRequest{
		Seeds:       seeds,
		Version:     version,
		UID:         uid,
		Username:    username,
		Password:    password,
		IsLowMemory: lowMemory == 0x01,
		CRCs:        crcs,
	}, nil
}
