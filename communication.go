package gommunication

import (
	"encoding/binary"
	"errors"
	"io"
)

type (
	Header struct {
		ID      uint16
		Version uint8
	}

	Message[BodyType any] struct {
		Header Header
		Body   BodyType
	}
)

const (
	messageStart = uint16(0xAAAA)
	messageEnd   = uint16(0xFFFF)

	// Missing start of message
	MissingSOM = "not at the start of a valid message"
	// Missing end of message
	MissingEOM = "there are bytes left on the message"
)

var (
	endianness = binary.LittleEndian
)

func (h *Header) FromStream(buf io.Reader) (err error) {
	var flag uint16
	err = binary.Read(buf, endianness, &flag)
	if err != nil {
		return
	}
	if flag != messageStart {
		return errors.New(MissingSOM)
	}

	return binary.Read(buf, endianness, h)
}

func (h *Header) ToStream(buf io.Writer) (err error) {
	err = binary.Write(buf, endianness, messageStart)
	if err != nil {
		return
	}

	return binary.Write(buf, endianness, h)
}

func (m *Message[BodyType]) FromStream(buf io.Reader) (err error) {
	// Process header
	err = m.Header.FromStream(buf)
	if err != nil {
		return
	}

	// Process body
	err = Deserialize[BodyType](buf, &m.Body)
	if err != nil {
		return
	}

	// Process footer
	var flag uint16
	err = binary.Read(buf, endianness, &flag)
	if err != nil {
		return
	}
	if flag != messageEnd {
		return errors.New(MissingEOM)
	}

	return
}

func (m *Message[BodyType]) ToStream(buf io.Writer) (err error) {
	// Write header
	err = m.Header.ToStream(buf)
	if err != nil {
		return
	}

	// Write body
	err = Serialize[BodyType](buf, m.Body)
	if err != nil {
		return
	}

	// Write footer and return
	return binary.Write(buf, endianness, messageEnd)
}
