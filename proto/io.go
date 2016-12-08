package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type UncompressedPacket struct {
	ID        PVarInt
	Component PacketComponentType
}

type CompressedPacket struct {
	DataLength PVarInt
	ID         PVarInt
	Component  PacketComponentType
}

func CreateWritablePacket(compressed bool, component PacketComponentWriteableType) PacketWriteable {
	return createPacket(compressed, component).(PacketWriteable)
}

func CreateReadablePacket(compressed bool, component PacketComponentReadableType) PacketReadable {
	return createPacket(compressed, component).(PacketReadable)
}

func createPacket(compressed bool, component PacketComponentType) Packet {
	if compressed {
		return &CompressedPacket{Component: component, ID: PVarInt(component.Id())}
	} else {
		return &UncompressedPacket{Component: component, ID: PVarInt(component.Id())}
	}
}

func (packet *CompressedPacket) GetId() int {
	return int(packet.ID)
}

func (packet *CompressedPacket) GetComponent() PacketComponentType {
	return packet.Component
}

func (packet *UncompressedPacket) GetId() int {
	return int(packet.ID)
}

func (packet *UncompressedPacket) GetComponent() PacketComponentType {
	return packet.Component
}

func (packet *UncompressedPacket) Write(destination *bytes.Buffer) (numWritten int, err error) {
	//write the length
	numWritten, err = packet.ID.Write(destination)
	if err != nil {
		return
	}

	//and the data (ID + packet)
	n, err := packet.Component.(PacketComponentWriteableType).Write(destination)
	if err != nil {
		return
	}
	numWritten = numWritten + n
	return
}

func (packet *UncompressedPacket) Read(source *bytes.Buffer) (numRead int, err error) {
	numRead, err = packet.ID.Read(source)
	if err != nil {
		return
	}
	n, err := packet.Component.(PacketComponentReadableType).Read(source)
	if err != nil {
		return
	}
	numRead = numRead + n
	return
}

func (packet *CompressedPacket) Read(source *bytes.Buffer) (numRead int, err error) {
	panic("Not implemented")
}

func (packet *CompressedPacket) Write(source *bytes.Buffer) (numWrite int, err error) {
	panic("Not implemented")
}

func (val *PVarInt) Write(buffer *bytes.Buffer) (numWritten int, err error) {
	return writeVarIntLong(uint64(*val), buffer)
}

func (val *PVarLong) Write(buffer *bytes.Buffer) (numWritten int, err error) {
	return writeVarIntLong(uint64(*val), buffer)
}

func writeVarIntLong(value uint64, buffer *bytes.Buffer) (numWritten int, err error) {
	buf := make([]byte, 12)
	numWritten = binary.PutUvarint(buf, value)
	buf = buf[:numWritten]
	buffer.Write(buf)
	return
}

func (val *PVarInt) Read(buffer *bytes.Buffer) (numRead int, err error) {
	var result uint64
	defer func() {
		*val = PVarInt(result)
	}()
	return readVarIntLong(buffer, &result, 5)
}

func (val *PVarLong) Read(buffer *bytes.Buffer) (numRead int, err error) {
	var result uint64
	defer func() {
		*val = PVarLong(result)
	}()
	return readVarIntLong(buffer, &result, 10)
}

func readVarIntLong(buffer *bytes.Buffer, dest *uint64, maxLen uint) (numRead int, err error) {
	start := buffer.Len()
	*dest, err = binary.ReadUvarint(buffer)
	numRead = start - buffer.Len()
	return
}

func (val *PString) Write(buffer *bytes.Buffer) (numWritten int, err error) {
	length := PVarInt(len(*val))
	n, err := (&length).Write(buffer)
	if err != nil {
		return
	}
	numWritten = numWritten + n

	n, err = buffer.WriteString(string(*val))
	if err == nil {
		numWritten = numWritten + n
	}
	return
}

func (val *PString) Read(buffer *bytes.Buffer) (numRead int, err error) {
	var rawLength PVarInt
	n, err := (&rawLength).Read(buffer)
	if err != nil {
		return
	}
	numRead = numRead + n

	var result string
	bytes := make([]byte, uint32(rawLength))
	n, err = io.ReadFull(buffer, bytes)
	if err != nil {
		return
	}
	numRead = numRead + n
	result = string(bytes)

	*val = PString(result)
	return
}

func (val *PUShort) Write(buffer *bytes.Buffer) (numWritten int, err error) {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, uint16(*val))
	return buffer.Write(bytes)
}

func (val *PUShort) Read(buffer *bytes.Buffer) (numRead int, err error) {
	bytes := make([]byte, 2)
	numRead, err = buffer.Read(bytes)
	if err != nil {
		return
	}

	*val = PUShort(binary.BigEndian.Uint16(bytes))
	return
}

func (val *PLong) Write(buffer *bytes.Buffer) (numWritten int, err error) {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(*val))
	return buffer.Write(bytes)
}

func (val *PLong) Read(buffer *bytes.Buffer) (numRead int, err error) {
	bytes := make([]byte, 8)
	numRead, err = buffer.Read(bytes)

	if numRead != len(bytes) {
		err = fmt.Errorf("invalid number of bytes read: %d", numRead)
	}

	if err != nil {
		return
	}

	*val = PLong(binary.BigEndian.Uint64(bytes))
	return
}
