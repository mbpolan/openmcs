package response

import "github.com/mbpolan/openmcs/internal/network"

const SetInterfaceSettingResponseHeader byte = 0x57

type SetInterfaceSettingResponse struct {
	SettingID int
	Value     int
}

// Write writes the contents of the message to a stream.
func (p *SetInterfaceSettingResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SetInterfaceSettingResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the setting id
	err = w.WriteUint16LE(uint16(p.SettingID))
	if err != nil {
		return err
	}

	// write 4 bytes for the value using a mixed byte format
	v := (p.Value&0x00FF)<<16 | (p.Value&0xFF00)>>16
	err = w.WriteUint32(uint32(v))
	if err != nil {
		return err
	}

	return nil
}
