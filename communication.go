package gommunication

import (
	"encoding/binary"
	"errors"
	"io"
)

// TODO ADD AUTOMATIC VERSIONING BASED ON STRUCT TAGS AND HEADER

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

	messageStart = uint16(0xAAAA)
	messageEnd   = uint16(0xFFFF)
)

var (
	// MissingSOH is the error returned when reading a header and failing.
	// Might mean more errors will popup if looking for more headers, but
	// isn't the end of the world.
	MissingSOH = errors.New("not at the start of a header")
	// MissingEOH is the error returned when there are bytes left in the header.
	// Unless you're doing something wrong, this should never happen.
	MissingEOH = errors.New("there are bytes left on the header")

	// MissingSOM is the error returned when failing to parse the start of a message.
	// If this fails FlushMessage can be used, or try parsing headers until one is properly found.
	MissingSOM = errors.New("not at the start of a valid message")
	// MissingEOM is the error returned when a message wasn't fully parsed.
	// It is possible to use FlushMessage to recover from this error. Can fail.
	MissingEOM = errors.New("there are bytes left on the message")
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
		return MissingSOH
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
		return MissingEOH
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
		err = MissingSOM
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
		err = MissingEOM
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

// FlushMessage attemps to clean the reader.
// Might not work even if no errors are returned.
// Always be weary after any read fails.
func FlushMessage(buf io.Reader) (err error) {
	var countEOM = 0
	for {
		var buffer = make([]byte, 1)
		_, err = buf.Read(buffer)
		if err != nil {
			return
		}
		if buffer[0] == 0xFF {
			countEOM++
		} else {
			countEOM = 0
		}
		if countEOM == 2 {
			return nil
		}
	}
}
