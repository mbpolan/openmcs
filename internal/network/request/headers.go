package request

// various client-sent requests have no payloads and therefore can be represented as just packet headers

// KeepAliveRequestHeader is sent by the client to maintain connectivity.
const KeepAliveRequestHeader byte = 0x00

// PlayerIdleRequestHeader is sent by the client to indicate the player has not interacted with the game in some time.
const PlayerIdleRequestHeader byte = 0xCA

// RegionLoadedRequestHeader is sent by the client to confirm that a map region has been loaded.
const RegionLoadedRequestHeader byte = 0x79
