package proto

import (
	"bytes"
)

func (packet *HandshakePacket) Write(dest *bytes.Buffer) (n int, err error) {
	nw, err := packet.ProtocolVersion.Write(dest)
	if err != nil {
		return
	}
	n = n + nw
	nw, err = packet.ServerAddress.Write(dest)
	if err != nil {
		return
	}
	n = n + nw
	nw, err = packet.ServerPort.Write(dest)
	if err != nil {
		return
	}
	n = n + nw
	nw, err = packet.PNextState.Write(dest)
	if err == nil {
		n = n + nw
	}
	return
}

func (packet *StatusRequest) Write(dest *bytes.Buffer) (n int, err error) {
	return 0, nil
}

func (packet *StatusResponse) Read(source *bytes.Buffer) (n int, err error) {
	return (&packet.JSONResponse).Read(source)
}

func (packet *ClientPing) Write(dest *bytes.Buffer) (n int, err error) {
	return (&packet.Payload).Write(dest)
}

func (packet *ServerPong) Read(source *bytes.Buffer) (n int, err error) {
	return (&packet.Payload).Read(source)
}
