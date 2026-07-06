package parser

type MessageHeader struct {
	Timestamp int
	TimestampDelta int // 최근에 들어온 Delta값
	Length int
	TypeId int
	StreamId int
}

type MessageHeaders = map[int]MessageHeader

func NewMessageHeader(timestamp int, length int, typeId int, streamId int) *MessageHeader {
	return &MessageHeader{
		Timestamp: timestamp,
		TimestampDelta: 0,
		Length: length,
		TypeId: typeId,
		StreamId: streamId,
	}
}