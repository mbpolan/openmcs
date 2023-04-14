package response

import "github.com/mbpolan/openmcs/internal/network"

type Response interface {
	Write(w *network.ProtocolWriter) error
}
