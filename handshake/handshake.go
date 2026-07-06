package handshake

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"slices"
	"time"
)

type Handshake struct {
	c net.Conn
	c1 []byte
	s1 []byte
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

	buf := append(s0[:], s1...)

	_, err = c.Write(buf)
	if err != nil {
		fmt.Println("Error while write: ", err)
		return err
	}

	if err := h.CheckC2(); err != nil {
		fmt.Println("Invalid C2 checked: ", err)
		return err
	}

	s2 := h.GenerateS2()

	_, err = c.Write(s2)
	if err != nil {
		fmt.Println("Error while write: ", err)
		return err
	}	

	fmt.Println("Handhsake Completed")

	return nil
}

func (h *Handshake) CheckC2() error {
	c2, err := h.ReadC2()
	if err != nil {
		return err
	}

	if slices.Equal(c2[0:4], h.s1[0:4]) && slices.Equal(c2[8:], h.s1[8:]) {
		return nil
	}

	return ErrInvalidC2
}

func (h *Handshake) ReadC2() ([]byte, error) {
	c2 := make([]byte, 1536)			
	_, err := io.ReadFull(h.c, c2);
	if err != nil {
		return nil, err
	}

	return c2, nil
}

func (h *Handshake) GenerateS2() []byte {
	s2 := make([]byte, len(h.c1))
	copy(s2, h.c1)

	copy(s2[4:8], h.s1[0:4])

	return s2
}

func (h *Handshake) GenerateS0S1() ([1]byte, []byte, error) {
	var version = [1]byte{3}

	buf := make([]byte, 1536)
	if _, err := rand.Read(buf[8:]); err != nil {
		return version, nil, err
	}

	timestamp := uint32(time.Now().UnixMilli())

	binary.BigEndian.PutUint32(buf[0:], timestamp)
	binary.BigEndian.PutUint32(buf[4:], 0)

	h.s1 = buf

	return version, buf, nil
}

func (h *Handshake) ReadC0C1() ([1]byte, []byte, error) {
	var version [1]byte

	c1 := make([]byte, 1536)

	_, err := io.ReadFull(h.c, version[:])
	if err != nil {
		return version, nil, err
	}

	_, err = io.ReadFull(h.c, c1[:])
	if err != nil {
		return version, nil, err
	}

	h.c1 = c1

	return version, c1, nil
}