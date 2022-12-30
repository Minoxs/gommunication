package gommunication

import (
	"encoding/binary"
	"errors"
	"io"
)

type (
	// Header is a helper struct that is appended before any
	// serialized data to add some additional information that
	// is used when processing a Message
	Header struct {
		ID      uint16
		Version uint8
	}

	// Message is the main struct used to serialize structured data.
	// Header is used to identify what kind of data is being saved
	// with support for versioning. Look for Serialize if you
	// only need serialization.
	Message[BodyType any] struct {
		Header Header
		Body   BodyType
	}
)

const (
	headerStart = uint8(0xBB)
	headerEnd   = uint8(0xCC)

	// Missing start of header
	MissingSOH = "not at the start of a header"
	// Missing end of header
	MissingEOH = "there are bytes left on the header"

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
	var flag uint8
	err = binary.Read(buf, endianness, &flag)
	if err != nil {
		return
	}
	if flag != headerStart {
		return errors.New(MissingSOH)
	}

	err = binary.Read(buf, endianness, h)
	if err != nil {
		return
	}

	err = binary.Read(buf, endianness, &flag)
	if err != nil {
		return
	}
	if flag != headerEnd {
		return errors.New(MissingEOH)
	}

	return nil
}

func (h *Header) ToStream(buf io.Writer) (err error) {
	err = binary.Write(buf, endianness, headerStart)
	if err != nil {
		return
	}

	err = binary.Write(buf, endianness, h)
	if err != nil {
		return
	}

	return binary.Write(buf, endianness, headerEnd)
}

func (m *Message[BodyType]) FromStream(buf io.Reader) (err error) {
	// Only process header if it wasn't already
	if m.Header == (Header{}) {
		err = m.Header.FromStream(buf)
		if err != nil {
			return
		}
	}

	m.Body, err = ReadBody[BodyType](buf)

	return
}

func (m *Message[BodyType]) ToStream(buf io.Writer) (err error) {
	// Write header
	err = m.Header.ToStream(buf)
	if err != nil {
		return
	}

	return WriteBody[BodyType](buf, m.Body)
}

func ReadBody[BodyType any](buf io.Reader) (res BodyType, err error) {
	var flag uint16
	err = binary.Read(buf, endianness, &flag)
	if err != nil {
	}
	if flag != messageStart {
		err = errors.New(MissingSOM)
		return
	}

	err = Deserialize[BodyType](buf, &res)
	if err != nil {
		return
	}

	err = binary.Read(buf, endianness, &flag)
	if err != nil {
		return
	}
	if flag != messageEnd {
		err = errors.New(MissingEOM)
		return
	}

	return
}

func WriteBody[BodyType any](buf io.Writer, body BodyType) (err error) {
	err = binary.Write(buf, endianness, messageStart)
	if err != nil {
		return
	}

	err = Serialize[BodyType](buf, body)
	if err != nil {
		return
	}

	return binary.Write(buf, endianness, messageEnd)
}
