package parser

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var ErrLastHeaderNotExists = errors.New("Last header does not exists")

type Parser struct {
	c net.Conn
	fmt uint8
	csid int
	lastHeaders MessageHeaders
}

func NewParser(c net.Conn) *Parser {
	return &Parser{
		c: c,
		lastHeaders: make(MessageHeaders),
	}
}

func (p *Parser) Load() error {
	if err := p.getBasicHeader(); err != nil {
		return fmt.Errorf("Parser: %w", err)
	} 

	if err := p.getChunkMessageHeader(); err != nil {
		return fmt.Errorf("Parser: %w", err)
	}

	return nil
}

func (p *Parser) getChunkMessageHeader() error {	
	// h := NewMessageHeader(0, 0, 0, 0)
	switch p.fmt {
	case 0:
		b := make([]byte, 11)
		if _, err := io.ReadFull(p.c, b); err != nil {
			return fmt.Errorf("Reading Chunk message header: %w", err)
		}

		timestamp := int(binary.BigEndian.Uint32([]byte{0, b[0], b[1], b[2]}))
		mLength := int(binary.BigEndian.Uint32([]byte{0, b[3], b[4], b[5]}))
		mTypeId := int(b[6])
		mStreamId := int(binary.LittleEndian.Uint32(b[7:11]))

		if timestamp >= 0xFFFFFF {
			var err error = nil
			timestamp, err = p.getExtraTimestamp()
			if err != nil {
				return err
			}
		}

		p.lastHeaders[p.csid] = *NewMessageHeader(timestamp, mLength, mTypeId, mStreamId)
	case 1:
		b := make([]byte, 7)
		if _, err := io.ReadFull(p.c, b); err != nil {
			return fmt.Errorf("Reading Chunk message header: %w", err)
		}

		timestampDelta := int(binary.BigEndian.Uint32([]byte{0, b[0], b[1], b[2]}))
		mLength := int(binary.BigEndian.Uint32([]byte{0, b[3], b[4], b[5]}))
		mTypeId := int(b[6])

		if timestampDelta >= 0xFFFFFF {
			var err error
			timestampDelta, err = p.getExtraTimestamp()
			if err != nil { return err }
		}

		value, exists := p.lastHeaders[p.csid]
	
		if !exists {
			return fmt.Errorf("Reading Chunk message header: %w", ErrLastHeaderNotExists)
		}

		value.Timestamp += timestampDelta
		value.TimestampDelta = timestampDelta
		value.Length = mLength
		value.TypeId = mTypeId

		p.lastHeaders[p.csid] = value
	case 2:
		b := make([]byte, 3)
		if _, err := io.ReadFull(p.c, b); err != nil {
			return fmt.Errorf("Reading Chunk message header: %w", err)
		}

		timestampDelta := int(binary.BigEndian.Uint32([]byte{0, b[0], b[1], b[2]}))

		if timestampDelta >= 0xFFFFFF {
			var err error
			timestampDelta, err = p.getExtraTimestamp()
			if err != nil { return err }
		}

		value, exists := p.lastHeaders[p.csid]

		if !exists {
			return fmt.Errorf("Reading Chunk message header: %w", ErrLastHeaderNotExists)
		}

		value.Timestamp += timestampDelta
		value.TimestampDelta = timestampDelta

		p.lastHeaders[p.csid] = value
	case 3:
		// TODO: Case 3 부분 명세서 다시 보고 구현하기
		value, exists := p.lastHeaders[p.csid]

		if !exists {
			return fmt.Errorf("Reading Chunk message header: %w", ErrLastHeaderNotExists)
		}

		timestampDelta := value.TimestampDelta
		if timestampDelta >= 0xFFFFFF {
			var err error
			timestampDelta, err = p.getExtraTimestamp()
			if err != nil { return err }
		}

		value.Timestamp += timestampDelta

		p.lastHeaders[p.csid] = value
	}

	return nil
}

func (p *Parser) getExtraTimestamp() (int, error) {
	ext := make([]byte, 4)
  if _, err := io.ReadFull(p.c, ext); err != nil {
    return 0, fmt.Errorf("Reading Extended Timestamp: %w", err)
  }
  timestamp := int(binary.BigEndian.Uint32(ext))

	return timestamp, nil
}

func (p *Parser) getBasicHeader() error {
	b_first := make([]byte, 1)	

	if _, err := io.ReadFull(p.c, b_first); err != nil {
		return fmt.Errorf("Reading Basic header: %w", err)
	}

	first := b_first[0]

	p.fmt = first >> 6

	csid := int(first & 0x3F)

	switch csid {
	case 0:
		ext := make([]byte, 1)	
		if _, err := io.ReadFull(p.c, ext); err != nil {
			return fmt.Errorf("Reading Basic header extension: %w", err)
		}

		csid = int(ext[0] + 64)
	case 1:
		ext := make([]byte, 2)	
		if _, err := io.ReadFull(p.c, ext); err != nil {
			return fmt.Errorf("Reading Basic header extension: %w", err)
		}

		csid = int(ext[1]) * 256 + int(ext[0]) + 64
	}

	p.csid = csid

	return nil
}