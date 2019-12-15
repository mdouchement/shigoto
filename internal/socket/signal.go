package socket

// SignalReload is the event for reloading all the Baito.
var SignalReload = Signal("reload")

// A Signal is an event sent through a socket.
type Signal []byte
