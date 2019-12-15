package socket

import (
	"bufio"
	"bytes"
	"net"

	"github.com/pkg/errors"
)

const (
	// ControlCharacter defines the end of socket communication.
	ControlCharacter byte = '\a'
	// ControlCharacterString defines the end of socket communication.
	ControlCharacterString = "\a"
)

// Request dials the socket with the given event.
func Request(socket string, event []byte) ([]byte, error) {
	c, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socket, Net: "unix"})
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
func Listen(socket string, handler func(event []byte) []byte) error {
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		return errors.Wrap(err, "could not open socket")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
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
