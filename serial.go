package serialpacket

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/tarm/serial"
)

// Frame Definition
// 	| LENGTH (1B) | DATA (LENGTH) |
// Max Packet Size: 254
const (
	HEADER_SIZE = 1
	MTU         = 254
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// header definition
type rawHeader [HEADER_SIZE]byte

func (h rawHeader) Length() int { return int(h[0]) }

// SerialPacketAddr is the address definition in net.Addr
type SerialPacketAddr struct{ name string }

func (addr *SerialPacketAddr) Network() string { return "serial" }
func (addr *SerialPacketAddr) String() string  { return "serial://" + addr.name }

// NewSerialPacketAddr creates an address with given name and port
func NewSerialPacketAddr(name string) *SerialPacketAddr {
	addr := new(SerialPacketAddr)
	addr.name = name
	return addr
}

// Conn is the packet connection definition for a serial connection
type Conn struct {
	port   *serial.Port
	addr   *SerialPacketAddr
	raddr  *SerialPacketAddr
	header rawHeader
}

func (conn *Conn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	// read full header
	n, err = io.ReadFull(conn.port, conn.header[:])
	if err != nil {
		return 0, nil, err
	}

	sz := conn.header.Length()
	if len(p) < sz {
		return 0, addr, fmt.Errorf("buffer size exceeded:need %v, given %v", sz, len(p))
	}

	// read full body
	n, err = io.ReadFull(conn.port, p[:sz])
	if err != nil {
		return n, addr, err
	}

	return n, addr, nil
}

func (conn *Conn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	if len(p) > MTU {
		return 0, fmt.Errorf("packet too large(MTU:%v) actual %v", MTU, len(p))
	}

	packet := make([]byte, HEADER_SIZE+len(p))
	header := packet[:HEADER_SIZE]
	data := packet[HEADER_SIZE:]
	header[0] = byte(len(p))
	copy(data, p)

	// write full packet until error
	written := 0
	for len(packet) > 0 {
		n, err = conn.port.Write(packet)
		if err != nil {
			return written, err
		} else {
			written += n
			packet = packet[written:]
		}
	}

	return written - HEADER_SIZE, nil
}

func (conn *Conn) Close() error                       { return conn.port.Close() }
func (conn *Conn) LocalAddr() net.Addr                { return conn.addr }
func (conn *Conn) SetDeadline(t time.Time) error      { return ErrNotImplemented }
func (conn *Conn) SetReadDeadline(t time.Time) error  { return ErrNotImplemented }
func (conn *Conn) SetWriteDeadline(t time.Time) error { return ErrNotImplemented }

// NewConn creates a net.PacketConn on a serial line
func NewConn(port *serial.Port) (*Conn, error) {
	c := new(Conn)
	c.port = port
	c.addr = NewSerialPacketAddr("local")
	c.raddr = NewSerialPacketAddr("remote")
	return c, nil
}