package frames

import (
	"bytes"

	"github.com/lucas-clemente/quic-go/protocol"
	"github.com/lucas-clemente/quic-go/utils"

	"errors"
)

var (
	errInvalidLenByte = errors.New("PLUSFeedbackFrame: Invalid len byte!")
	errUnexpectedEndOfData = errors.New("PLUSFeedbackFrame: Unexpected end of data!")
	errInvalidFrameType = errors.New("PLUSFeedbackFrame: Invalid fame type!")
	plusFeedbackFrameType byte = 0x08
)

// A BlockedFrame in QUIC
type PLUSFeedbackFrame struct {
	StreamID protocol.StreamID
	data []byte
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

	// read stream id
	sid, err := utils.ReadUint32(r)
	if err != nil {
		return nil, err
	}

	frame.StreamID = protocol.StreamID(sid)

	// read the len byte
	lenByte, err := r.ReadByte()

	// Since PCF value is limited to 64 bytes pcf feedback 
    // can also be 64 bytes max.
	if lenByte >= 64 {
		return nil, errInvalidLenByte
	}

	data := make([]byte, lenByte)

	n, err := r.Read(data)

	if n != int(lenByte) {
		return nil, errUnexpectedEndOfData
	}

	frame.data = data

	return frame, nil
}

//Write writes a PLUSFeedbackFrame frame
func (f *PLUSFeedbackFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	// Write type byte
	err := b.WriteByte(plusFeedbackFrameType)

	if err != nil {
		return err
	}

	// Write streamID
	utils.WriteUint32(b, uint32(f.StreamID))

	// Write len byte
	if len(f.data) >= 64 {
		return errInvalidLenByte
	}

	err = b.WriteByte(byte(len(f.data)))

	if err != nil {
		return err
	}

	n, err := b.Write(f.data)

	if err != nil {
		return err
	}

	if n != len(f.data) {
		return errors.New("PLUSFeedbackFrame: Write did not write enough bytes!")
	}

	return nil
}
