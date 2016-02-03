package vxlan

import (
	"bytes"
	"io"
	"testing"

	"github.com/mdlayher/ethernet"
)

func TestFrameMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		f    *Frame
		b    []byte
		err  error
	}{
		{
			desc: "Frame with VNI value too large",
			f: &Frame{
				VNI: MaxVNI + 1,
			},
			err: ErrInvalidVNI,
		},
		{
			desc: "Frame with Ethernet frame with VLAN too large",
			f: &Frame{
				Ethernet: &ethernet.Frame{
					VLAN: []*ethernet.VLAN{{
						ID: ethernet.VLANMax + 1,
					}},
				},
			},
			err: ethernet.ErrInvalidVLAN,
		},
		{
			desc: "Frame with VNI 1",
			f: &Frame{
				VNI:      1,
				Ethernet: &ethernet.Frame{},
			},
			b: append([]byte{
				0x08, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x01, 0x00,
			}, ethernetFrame(t)...),
		},
		{
			desc: "Frame with VNI Max",
			f: &Frame{
				VNI:      MaxVNI,
				Ethernet: &ethernet.Frame{},
			},
			b: append([]byte{
				0x08, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0x00,
			}, ethernetFrame(t)...),
		},
	}

	for i, tt := range tests {
		t.Logf("[%02d] test %q", i, tt.desc)

		b, err := tt.f.MarshalBinary()
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error: %v != %v",
					want, got)
			}

			continue
		}

		if want, got := tt.b, b; !bytes.Equal(want, got) {
			t.Fatalf("unexpected Frame bytes:\n- want: %v\n-  got: %v",
				want, got)
		}
	}
}

func TestFrameUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		b    []byte
		f    *Frame
		err  error
	}{
		{
			desc: "Frame too short",
			b:    []byte{0x00},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "Frame with I flag not set",
			b: append([]byte{
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x01, 0x00,
			}, ethernetFrame(t)...),
			err: ErrInvalidFrame,
		},
		{
			desc: "Frame with invalid Ethernet frame check sequence",
			b: func() []byte {
				b := append([]byte{
					0x08, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00,
				}, ethernetFrame(t)...)

				// Break FCS
				b[len(b)-1] = 0x00
				return b
			}(),
			err: ethernet.ErrInvalidFCS,
		},
		{
			desc: "Frame with VNI 1",
			b: append([]byte{
				0x08, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x01, 0x00,
			}, ethernetFrame(t)...),
			f: &Frame{
				VNI:      1,
				Ethernet: &ethernet.Frame{},
			},
		},
		{
			desc: "Frame with VNI Max",
			b: append([]byte{
				0x08, 0x00, 0x00, 0x00,
				0xff, 0xff, 0xff, 0x00,
			}, ethernetFrame(t)...),
			f: &Frame{
				VNI:      MaxVNI,
				Ethernet: &ethernet.Frame{},
			},
		},
	}

	for i, tt := range tests {
		t.Logf("[%02d] test %q", i, tt.desc)

		f := new(Frame)
		if err := f.UnmarshalBinary(tt.b); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error: %v != %v",
					want, got)
			}

			continue
		}

		fb, err := f.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if want, got := tt.b, fb; !bytes.Equal(want, got) {
			t.Fatalf("unexpected Frame bytes:\n- want: %v\n-  got: %v",
				want, got)
		}
	}
}

func ethernetFrame(t *testing.T) []byte {
	f := new(ethernet.Frame)
	fb, err := f.MarshalFCS()
	if err != nil {
		t.Fatalf("failed to marshal Ethernet frame: %v", err)
	}

	return fb
}
