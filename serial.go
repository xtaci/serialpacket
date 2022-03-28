package serialpacket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/tarm/serial"
)

// Frame Definition
// 	|MAGIC(4B) | LENGTH (2B) | DATA (LENGTH) |
// Max Packet Size: 240
const (
	HEADER_SIZE   = 2
	MAGIC_SIZE    = 4
	MTU           = 1000
	MAX_DATA_SIZE = MTU - HEADER_SIZE - MAGIC_SIZE
)

var (
	MagicBytes = []byte{0xFF, 0x00, 0xAA, 0x55}
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// header definition
type rawHeader [HEADER_SIZE]byte

func (h rawHeader) Length() int { return int(binary.LittleEndian.Uint16(h[:])) }

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
	port     *serial.Port
	addr     *SerialPacketAddr
	raddr    *SerialPacketAddr
	header   rawHeader
	magicPos int
	magic    [4]byte
}

func (c *Conn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	// 0xFF00AA55 to sync frame
	for {
		n, err = c.port.Read(c.magic[c.magicPos : c.magicPos+1])
		if err != nil {
			return 0, nil, err
		}

		if c.magic[c.magicPos] != MagicBytes[c.magicPos] {
			c.magicPos = 0
			continue
		}

		if c.magicPos == 3 {
			c.magicPos = 0
			break
		}
		c.magicPos++
	}

	// read full header
	n, err = io.ReadFull(c.port, c.header[:])
	if err != nil {
		return 0, nil, err
	}
	// log.Println("header:", c.header)

	sz := c.header.Length()
	if len(p) < sz {
		return 0, c.raddr, fmt.Errorf("buffer too small: need %v, given %v", sz, len(p))
	}

	// read full body
	n, err = io.ReadFull(c.port, p[:sz])
	if err != nil {
		return n, c.raddr, err
	}

	//log.Println("body:", sz, p[:sz])
	return n, c.raddr, nil
}

func (c *Conn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	if len(p) > MAX_DATA_SIZE {
		return 0, fmt.Errorf("packet too large(MAX_DATA_SIZE:%v) actual %v", MAX_DATA_SIZE, len(p))
	}

	packet := make([]byte, MAGIC_SIZE+HEADER_SIZE+len(p))
	magic := packet[:]
	copy(magic, MagicBytes)

	header := magic[MAGIC_SIZE:]
	binary.LittleEndian.PutUint16(header, uint16(len(p)))

	data := header[HEADER_SIZE:]
	copy(data, p)

	// write full packet until error
	written := 0
	for len(packet) > 0 {
		n, err = c.port.Write(packet)
		if err != nil {
			return written, err
		} else {
			written += n
			packet = packet[written:]
		}
	}

	return written - HEADER_SIZE - MAGIC_SIZE, nil
}

func (c *Conn) Close() error                       { return c.port.Close() }
func (c *Conn) LocalAddr() net.Addr                { return c.addr }
func (c *Conn) SetDeadline(t time.Time) error      { return ErrNotImplemented }
func (c *Conn) SetReadDeadline(t time.Time) error  { return ErrNotImplemented }
func (c *Conn) SetWriteDeadline(t time.Time) error { return ErrNotImplemented }

// NewConn creates a net.PacketConn on a serial line
func NewConn(port *serial.Port) (*Conn, error) {
	c := new(Conn)
	c.port = port
	c.addr = NewSerialPacketAddr("local")
	c.raddr = NewSerialPacketAddr("remote")
	return c, nil
}
