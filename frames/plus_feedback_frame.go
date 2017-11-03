package frames

import (
	"bytes"

	"github.com/lucas-clemente/quic-go/protocol"

	"errors"
)

var (
	errInvalidLenByte = errors.New("PLUSFeedbackFrame: Invalid len byte!")
	errUnexpectedEndOfData = errors.New("PLUSFeedbackFrame: Unexpected end of data!")
	errInvalidFrameType = errors.New("PLUSFeedbackFrame: Invalid fame type!")
	plusFeedbackFrameType byte = 0x08
)

// A PLUSFeedbackFrame in QUIC
type PLUSFeedbackFrame struct {
	Data []byte
}

// MinLength of a written frame
// Returns the FULL length. Not if packet_packer/stream_framer can handle variably sized
// *control* frames.
func (f *PLUSFeedbackFrame) MinLength(version protocol.VersionNumber) (protocol.ByteCount, error) {
	return protocol.ByteCount(1 + len(f.Data)), nil
}

// ParsePLUSFeedbackFrame reads a pcf frame
func ParsePLUSFeedbackFrame(r *bytes.Reader) (*PLUSFeedbackFrame, error) {
	frame := &PLUSFeedbackFrame{}

	// read type byte
	typeByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if typeByte != plusFeedbackFrameType {
		return nil, errInvalidFrameType
	}

	// read the len byte
	lenByte, err := r.ReadByte()

	

	data := make([]byte, lenByte)

	n, err := r.Read(data)

	if n != int(lenByte) {
		return nil, errUnexpectedEndOfData
	}

	frame.Data = data

	return frame, nil
}

//Write writes a PLUSFeedbackFrame frame
func (f *PLUSFeedbackFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	// Write type byte
	err := b.WriteByte(plusFeedbackFrameType)

	if err != nil {
		return err
	}

	// Write len byte
	if len(f.Data) > 255 {
		return errInvalidLenByte
	}

	err = b.WriteByte(byte(len(f.Data)))

	if err != nil {
		return err
	}

	n, err := b.Write(f.Data)

	if err != nil {
		return err
	}

	if n != len(f.Data) {
		return errors.New("PLUSFeedbackFrame: Write did not write enough bytes!")
	}

	return nil
}
