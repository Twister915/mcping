package proto

import (
	"bytes"
)

type PVarInt int32
type PVarLong int64
type PString string
type PUShort uint16
type PLong int64

type PacketWriteableDataType interface {
	Write(*bytes.Buffer) (n int, err error)
}

type PacketReadableDataType interface {
	Read(*bytes.Buffer) (n int, err error)
}

type PacketComponentType interface {
	Id() int
}

type PacketComponentReadableType interface {
	PacketComponentType
	PacketReadableDataType
}

type PacketComponentWriteableType interface {
	PacketComponentType
	PacketWriteableDataType
}

type Packet interface {
	GetId() int
	GetComponent() PacketComponentType
}

type PacketWriteable interface {
	Packet
	PacketWriteableDataType
}

type PacketReadable interface {
	Packet
	PacketReadableDataType
}

type PacketPerformAction interface {
	OnTransmit(conn *ServerConnection)
}
