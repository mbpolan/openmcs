package response

import "github.com/mbpolan/openmcs/internal/network"

const PlayMusicResponseHeader byte = 0x4A

// PlayMusicNoneID represents that the client should not play any music.
const PlayMusicNoneID = 0x00FFFF

// PlayMusicResponse is sent by the server to instruct the client to start playing a music track.
type PlayMusicResponse struct {
	MusicID int
}

// Write writes the contents of the message to a stream.
func (p *PlayMusicResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(PlayMusicResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the music id
	err = w.WriteUint16LE(uint16(p.MusicID))
	if err != nil {
		return err
	}

	return nil
}
