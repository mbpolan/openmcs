package request

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const ReportRequestHeader byte = 0xDA

// ReportRequest is sent by the client when a player files an abuse report.
type ReportRequest struct {
	Username   string
	Reason     int
	EnableMute bool
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *ReportRequest) Read(r *network.ProtocolReader) error {
	// read 8 bytes containing an encoded player name
	name, err := r.Uint64()
	if err != nil {
		return err
	}

	// decode the player name
	username, err := util.DecodeName(name)
	if err != nil {
		return err
	}

	// read 1 byte for the report reason
	reason, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 1 byte for a flag if the player should be muted
	mute, err := r.Uint8()
	if err != nil {
		return err
	}

	p.Username = username
	p.Reason = int(reason)
	p.EnableMute = mute == 0x01
	return nil
}
