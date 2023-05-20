package response

import "github.com/mbpolan/openmcs/internal/network"

const SetInterfaceTextResponseHeader byte = 0x7E

// SetInterfaceTextResponse is sent by the server to update the text on an interface.
type SetInterfaceTextResponse struct {
	interfaceID int
	text        string
}

// NewSetInterfaceTextResponse creates a new response to update the text on an interface.
func NewSetInterfaceTextResponse(interfaceID int, text string) *SetInterfaceTextResponse {
	return &SetInterfaceTextResponse{
		interfaceID: interfaceID,
		text:        text,
	}
}

// Write writes the contents of the message to a stream.
func (p *SetInterfaceTextResponse) Write(w *network.ProtocolWriter) error {
	// use a buffered writer since we need to keep track of length of the data
	bw := network.NewBufferedWriter()

	// write the text string
	err := bw.WriteString(p.text)
	if err != nil {
		return err
	}

	// write 2 bytes for the interface id
	err = bw.WriteUint16Alt3(uint16(p.interfaceID))
	if err != nil {
		return err
	}

	// write packet header
	err = w.WriteUint8(SetInterfaceTextResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the length of the buffered data
	buf, err := bw.Buffer()
	err = w.WriteUint16(uint16(buf.Len()))
	if err != nil {
		return err
	}

	// finally write the payload itself
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
