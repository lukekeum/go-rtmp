package handshake

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type Handshake struct {
	c net.Conn
	c1 []byte
	timestamp uint32
}

func NewHandshake(c net.Conn) (*Handshake) {
	return &Handshake{
		c: c,
	}
}

func Connect(c net.Conn) error {
	h := NewHandshake(c)

	c0, _, err := h.ReadC0C1()
	if err != nil {
		fmt.Println("Error while ReadC0C1: ", err)
		c.Close()
		return err
	}
	if c0[0] != 3 {
		fmt.Println("Invalid Version")
		return ErrInvalidVersion
	}

	s0, s1, err := h.GenerateS0S1()
	if err != nil {
		fmt.Println("Error while generate random bytes")
		return err
	}

	buf := append(s0, s1...)

	_, err = c.Write(buf)
	if err != nil {
		fmt.Println("Error while write: ", err)
	}

	s2 := h.GenerateS2()

	_, err = c.Write(s2)
	if err != nil {
		fmt.Println("Error while write: ", err)
	}	

	return nil
}

func (h *Handshake) GenerateS2() []byte {
	var s2 []byte
	copy(s2, h.c1)
	binary.BigEndian.PutUint32(s2[4:], h.timestamp)	

	return s2
}

func (h *Handshake) GenerateS0S1() ([]byte, []byte, error) {
	buf := make([]byte, 1536)
	random := make([]byte, 1528)

	if _, err := rand.Read(random); err != nil {
		return []byte{3}, buf, err
	}

	timestamp := uint32(time.Now().UnixMilli())

	h.timestamp = timestamp

	binary.BigEndian.PutUint32(buf[0:], timestamp)
	binary.BigEndian.PutUint32(buf[4:], 0)

	copy(buf[8:], random)

	return []byte{3}, buf, nil
}

func (h *Handshake) ReadC0C1() ([1]byte, [1536]byte, error) {
	var version [1]byte
	var c1 [1536]byte

	_, err := io.ReadFull(h.c, version[:])
	if err != nil {
		return version, c1, err
	}

	_, err = io.ReadFull(h.c, c1[:])
	if err != nil {
		return version, c1, err
	}

	h.c1 = c1[:]

	return version, c1, nil
}