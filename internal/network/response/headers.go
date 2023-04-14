package response

// various server-sent responses have no payloads and therefore can be represented as packet headers

// DisconnectResponseHeader is sent by the server to tell the client that a player has been forcefully logged out.
const DisconnectResponseHeader byte = 0x6D
