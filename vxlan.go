// Package vxlan implements marshaling and unmarshaling of Virtual eXtensible
// Local Area Network (VXLAN) frames, as described in RFC 7348.
package vxlan

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/mdlayher/ethernet"
)

const (
	// MaxVNI is the maximum possible value for a VNI: the maximum value
	// of a 24-bit integer.
	MaxVNI = (1 << 24) - 1
)

var (
	// ErrInvalidFrame is returned when the reserved I bit is not set in
	// a byte slice when Frame.UnmarshalBinary is called.
	ErrInvalidFrame = errors.New("invalid Frame")

	// ErrInvalidVNI is returned when a Frame contains an invalid VNI,
	// and Frame.MarshalBinary is called.
	ErrInvalidVNI = errors.New("invalid VNI")
)

// A VNI is a 24-bit Virtual Network Identifier.  It is used to designate a
// VXLAN overlay network.  Use its Valid method to determine if a VNI contains
// a valid value.
type VNI uint32

// Valid determines if a VNI is a valid, 24-bit integer.
func (v VNI) Valid() bool {
	return v <= MaxVNI
}

// A Frame is an Virtual eXtensible Local Area Network (VXLAN) frame, as
// described in RFC 7348, Section 5.
//
// It contains a VNI used to designate an overlay network, and an embedded
// Ethernet frame which transports an arbitrary payload within the overlay
// network.
type Frame struct {
	VNI      VNI
	Ethernet *ethernet.Frame
}

// MarshalBinary allocates a byte slice and marshals a Frame into binary form.
//
// If a VNI value is invalid, ErrInvalidVNI will be returned.
func (f *Frame) MarshalBinary() ([]byte, error) {
	if !f.VNI.Valid() {
		return nil, ErrInvalidVNI
	}

	efb, err := f.Ethernet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)

	// I flag is always set to 1, all others are reserved
	b[0] |= 1 << 3

	binary.BigEndian.PutUint32(b[3:], uint32(f.VNI))

	// Ethernet frame bytes accompany VXLAN frame
	return append(b, efb...), nil
}

// UnmarshalBinary allocates a byte slice and marshals a Frame into binary form.
//
// If a VNI value is invalid, ErrInvalidVNI will be returned.
func (f *Frame) UnmarshalBinary(b []byte) error {
	// Need at least VXLAN frame and empty Ethernet frame
	if len(b) < 18 {
		return io.ErrUnexpectedEOF
	}

	// I flag must be set to 1.
	if (b[0] >> 3) != 1 {
		return ErrInvalidFrame
	}

	f.VNI = VNI(binary.BigEndian.Uint32(b[3:]))

	ef := new(ethernet.Frame)
	if err := ef.UnmarshalBinary(b[8:]); err != nil {
		return err
	}
	f.Ethernet = ef

	return nil
}
