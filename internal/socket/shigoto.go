package socket

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	// ControlCharacter defines the end of socket communication.
	ControlCharacter byte = '\a'
	// ControlCharacterString defines the end of socket communication.
	ControlCharacterString = "\a"
)

// A Socket is used to send and receive signals through socket.
type Socket struct {
	socket string
	close  func() error
}

// New returns a new Socket.
func New(socket string) *Socket {
	return &Socket{
		socket: socket,
	}
}

// Request dials the socket with the given event.
func (s *Socket) Request(event []byte) ([]byte, error) {
	c, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: s.socket, Net: "unix"})
	if err != nil {
		return nil, errors.Wrap(err, "could not dial the socket")
	}
	_, err = c.Write(append([]byte(event), ControlCharacter))
	if err != nil {
		return nil, errors.Wrap(err, "could not request the event")
	}

	reader := bufio.NewReader(c)
	payload, err := reader.ReadBytes(ControlCharacter)
	if err != nil {
		return nil, err
	}
	return bytes.Trim(payload, ControlCharacterString), nil
}

// Listen opens the socket and wait for incoming events.
func (s *Socket) Listen(handler func(event []byte) []byte) error {
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: s.socket, Net: "unix"})
	if err != nil {
		return errors.Wrap(err, "could not open socket")
	}
	s.close = ln.Close

	for {
		conn, err := ln.Accept()
		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return nil // Close has been trigged
			}
			return err
		}
		reader := bufio.NewReader(conn)
		if event, err := reader.ReadBytes(ControlCharacter); err == nil {
			event = bytes.Trim(event, ControlCharacterString)

			payload := handler(event)
			_, _ = conn.Write(append(payload, ControlCharacter))
		}
	}
}

// Close closes the listener if Listen as been called.
func (s *Socket) Close() error {
	defer os.Remove(s.socket)

	if s.close == nil {
		return nil
	}
	return s.close()
}
