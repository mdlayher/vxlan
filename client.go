package vxlan

import (
	"net"

	"golang.org/x/net/ipv4"
)

type Client struct {
	p *ipv4.PacketConn
	c net.PacketConn

	b []byte
}

func NewClient(ifi *net.Interface, group net.IP) (*Client, error) {
	c, err := net.ListenPacket("udp4", "0.0.0.0:8472")
	if err != nil {
		return nil, err
	}

	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(ifi, &net.UDPAddr{IP: group}); err != nil {
		return nil, err
	}

	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		return nil, err
	}

	client := &Client{
		p: p,
		c: c,
		b: make([]byte, ifi.MTU),
	}

	return client, nil
}

func (c *Client) Read() (*Frame, net.Addr, error) {
	n, _, src, err := c.p.ReadFrom(c.b)
	if err != nil {
		return nil, nil, err
	}

	f := new(Frame)
	err = f.UnmarshalBinary(c.b[:n])

	return f, src, err
}
