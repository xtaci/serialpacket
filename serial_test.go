package serialpacket

import (
	"crypto/sha1"
	"fmt"
	"os"
	"testing"

	"github.com/tarm/serial"
	kcp "github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

const (
	pass = "serialtest"
	salt = "asdklfalsdjfalkjdflakjdsf"
)

func TestConnection(t *testing.T) {
	port0 := os.Getenv("PORT0")
	port1 := os.Getenv("PORT1")
	if port0 == "" || port1 == "" {
		t.Skip("Skipping test because PORT0 or PORT1 environment variable is not set")
	}
	c0 := &serial.Config{Name: port0, Baud: 9600}
	c1 := &serial.Config{Name: port1, Baud: 9600}

	s1, err := serial.OpenPort(c0)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := serial.OpenPort(c1)
	if err != nil {
		t.Fatal(err)
	}

	// wrapp to net.PacketConn
	conn1, err := NewConn(s1)
	if err != nil {
		t.Fatal(err)
	}

	conn2, err := NewConn(s2)
	if err != nil {
		t.Fatal(err)
	}

	// crypt
	seed := pbkdf2.Key([]byte(pass), []byte(salt), 1024, 32, sha1.New)
	block1, _ := kcp.NewSalsa20BlockCrypt(seed)
	block2, _ := kcp.NewSalsa20BlockCrypt(seed)

	// listen
	l, err := kcp.ServeConn(block1, 1, 0, conn1)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			s, err := l.Accept()
			if err != nil {
				return
			}
			s.(*kcp.UDPSession).SetMtu(200)
			go handleEcho(s.(*kcp.UDPSession))
		}
	}()

	// client
	cli, err := kcp.NewConn("", block2, 1, 0, conn2)
	if err != nil {
		t.Fatal(err)
	}
	const N = 100
	buf := make([]byte, 10)
	for i := 0; i < N; i++ {
		msg := fmt.Sprintf("hello%v", i)
		cli.Write([]byte(msg))
		if n, err := cli.Read(buf); err == nil {
			if string(buf[:n]) != msg {
				t.Fail()
			}
		} else {
			panic(err)
		}
	}
	cli.Close()
}

func handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}
