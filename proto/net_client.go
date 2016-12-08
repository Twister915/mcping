package proto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type ProtocolState int

const (
	Handshaking ProtocolState = iota
	Play
	Status
	Login
)

type ServerConnection struct {
	packetsRead  chan PacketComponentReadableType
	packetsWrite chan PacketComponentWriteableType

	errors chan error

	protocolState      ProtocolState
	compressed, closed bool

	connection *net.TCPConn
	reader     *bufio.Reader

	mutex *sync.Mutex

	encryption struct {
	}
}

func Connect(addr string) (conn *ServerConnection, err error) {
	rawDialedConn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	conn = &ServerConnection{}
	conn.connection = rawDialedConn.(*net.TCPConn)
	conn.packetsRead = make(chan PacketComponentReadableType)
	conn.packetsWrite = make(chan PacketComponentWriteableType)
	conn.errors = make(chan error)
	conn.protocolState = Handshaking
	conn.mutex = &sync.Mutex{}
	conn.reader = bufio.NewReader(rawDialedConn)

	return
}

func (conn *ServerConnection) Close() {
	//do not unlock, we do not want another thread to ever get a lock on these again
	conn.mutex.Lock()

	conn.closed = true
	conn.compressed = false
	conn.connection.Close()
	close(conn.errors)
	close(conn.packetsRead)
	close(conn.packetsWrite)
}

func (conn *ServerConnection) Serve() {
	go func() {
		for !conn.closed {
			packet, data, err := conn.getPacket()
			if conn.closed {
				return
			}
			if err != nil {
				if err == io.EOF {
					continue
				}
				conn.errors <- err
				continue
			}
			debug(fmt.Sprintf("Read %d bytes for %d", len(data), packet.GetId()))
			buffer := bytes.NewBuffer(data)
			_, err = packet.Read(buffer)
			if err != nil {
				if err == io.EOF {
					continue
				} else {
					conn.errors <- err
					continue
				}
			}
			func() {
				conn.mutex.Lock()
				defer conn.mutex.Unlock()

				if !conn.closed {
					conn.packetsRead <- packet.GetComponent().(PacketComponentReadableType)
					conn.onPacket(packet.GetComponent())
				}
			}()
		}
	}()

	go func() {
		var buf, lenBuf bytes.Buffer
		for component := range conn.packetsWrite {
			packet := CreateWritablePacket(conn.compressed, component)
			_, err := packet.Write(&buf)
			if err != nil {
				conn.errors <- err
				continue
			}

			packed := buf.Bytes()
			len := PVarInt(len(packed))
			_, err = (&len).Write(&lenBuf)
			if err != nil {
				conn.errors <- err
				continue
			}

			lenBytes := lenBuf.Bytes()
			allBytes := make([]byte, int(len)+lenBuf.Len())
			copy(allBytes, lenBytes)
			copy(allBytes[lenBuf.Len():], packed)

			lenBuf.Reset()
			buf.Reset()

			func() {
				conn.mutex.Lock()
				defer conn.mutex.Unlock()

				debug(fmt.Sprintf("Write %o", allBytes))
				_, err = conn.connection.Write(allBytes)

				if err != nil {
					if err == io.EOF {
						return
					}
					conn.errors <- err
				} else {
					conn.onPacket(component)
				}
			}()
		}
	}()
}

func (conn *ServerConnection) onPacket(packet PacketComponentType) {
	if changer, ok := packet.(PacketPerformAction); ok {
		changer.OnTransmit(conn)
	}
}

func (conn *ServerConnection) getPacket() (packet PacketReadable, data []byte, err error) {
	//manual read var int thingy

	//first get length
	length, err := binary.ReadUvarint(conn.reader)
	if err != nil {
		return
	}
	rawBytes := make([]byte, int(length))
	n, err := io.ReadFull(conn.reader, rawBytes)
	if uint64(n) != length {
		panic("did not read enough bytes...")
	}

	var packetComponent PacketComponentType
	var id PVarInt
	_, err = (&id).Read(bytes.NewBuffer(rawBytes))

	switch conn.protocolState {
	case Handshaking:
		packetComponent = getPacketByIdHandshaking(int(id))
	case Status:

		packetComponent = getPacketByIdStatus(int(id))
	default:
		err = fmt.Errorf("No protocol state selected!")
		return
	}
	if packetComponent == nil {
		err = fmt.Errorf("Could not select packet for reading...")
		return
	}
	data = rawBytes
	packet = CreateReadablePacket(conn.compressed, packetComponent.(PacketComponentReadableType))
	return
}

func getPacketByIdHandshaking(id int) (packet PacketComponentType) {
	return
}

func getPacketByIdStatus(id int) (packet PacketComponentType) {
	switch id {
	case 0:
		packet = &StatusResponse{}
	case 1:
		packet = &ServerPong{}
	}
	return
}

func (conn *ServerConnection) IncomingPackets() <-chan PacketComponentReadableType {
	return conn.packetsRead
}

func (conn *ServerConnection) OutgoingPackets() chan<- PacketComponentWriteableType {
	return conn.packetsWrite
}

func (conn *ServerConnection) Errors() <-chan error {
	return conn.errors
}

func (conn *ServerConnection) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}
