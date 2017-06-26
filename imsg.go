// imsg.go - OpenBSD imsg RPC facility.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to imsg, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package imsg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
)

type Header struct {
	Type   uint32
	Length uint16
	Flags  uint16
	PeerID uint32
	PID    uint32
}

const (
	HeaderSize     = 4 + 2 + 2 + 4 + 4
	MaxMessageSize = 16384
)

func (h *Header) Marshal() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, h)
	return buf.Bytes()
}

func UnmarshalHeader(b []byte) (*Header, error) {
	hdr := &Header{}
	r := bytes.NewReader(b)
	err := binary.Read(r, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}
	return hdr, nil
}

type Conn struct {
	rwc io.ReadWriteCloser
	pid uint32
}

func NewConn(c net.Conn) (*Conn, error) {
	ic := &Conn{
		rwc: c,
		pid: uint32(os.Getpid()),
	}
	return ic, nil
}

func (ic *Conn) Send(typ uint32, flags uint16, peerid uint16, data []byte) error {
	if len(data) > MaxMessageSize {
		return errors.New("message is too large")
	}
	msglen := HeaderSize + len(data)
	b := make([]byte, 0, msglen)
	hdr := &Header{
		Type:   typ,
		Length: uint16(msglen),
		PID:    ic.pid,
	}
	b = append(b, hdr.Marshal()...)
	b = append(b, data...)
	_, err := ic.rwc.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (ic *Conn) Recv() (*Header, []byte, error) {
	var data []byte
	hb := make([]byte, HeaderSize)
	n, err := ic.rwc.Read(hb)
	if err != nil {
		return nil, nil, err
	}
	hdr, err := UnmarshalHeader(hb[:n])
	if err != nil {
		return nil, nil, err
	}
	if hdr.Length > 0 {
		datalen := int(hdr.Length) - HeaderSize
		data = make([]byte, datalen)
		_, err := io.ReadFull(ic.rwc, data)
		if err != nil {
			return nil, nil, err
		}
	}
	return hdr, data, nil
}
