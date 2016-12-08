package proto

//protocol=handshake; direction=serverbound; ID=0x00
type HandshakePacket struct {
	ProtocolVersion PVarInt
	ServerAddress   PString
	ServerPort      PUShort
	PNextState      PVarInt
}

func (packet *HandshakePacket) Id() int { return 0 }

func (packet *HandshakePacket) OnTransmit(conn *ServerConnection) {
	var state ProtocolState
	switch packet.PNextState {
	case 1:
		state = Status
	case 2:
		state = Login
	default:
		panic("Bad next state for Handshake Packet")
	}
	conn.protocolState = state
}

//protocol=status; direction=serverbound; ID=0x00
type StatusRequest struct{}

func (packet *StatusRequest) Id() int { return 0x00 }

//protocol=status; direction=clientbound; ID = 0x00
type StatusResponse struct {
	JSONResponse PString
}

func (packet *StatusResponse) Id() int { return 0x00 }

//protocol=status; direction=serverbound; ID= 0x01
type ClientPing struct {
	Payload PLong
}

func (packet *ClientPing) Id() int { return 0x01 }

//protocol=status; direction=clientbound; ID=0x01
type ServerPong struct {
	Payload PLong
}

func (packet *ServerPong) Id() int { return 0x01 }

//protocol=login; direction=serverbound; ID=0x00
type LoginStart struct {
	Username PString
}

func (packet *LoginStart) Id() int { return 0x00 }

//protocol=login; direction=serverbound; ID=0x01
type EncryptionResponse struct {
	SharedSecretLength PVarInt
	SharedServer       []byte
	VerifyTokenLength  PVarInt
	VerifyToken        []byte
}

func (packet *EncryptionResponse) Id() int { return 0x01 }

//protocol=login; direction=clientbound; ID=0x00
type Disconnect struct {
	Reason PString
}

func (packet *Disconnect) Id() int { return 0x00 }

//protocol=login; direction=clientbound; ID=0x01
type EncryptionRequest struct {
	ServerID          PString
	PublicKeyLength   PVarInt
	PublicKey         []byte
	VerifyTokenLength PVarInt
	VerifyToken       []byte
}

func (packet *EncryptionRequest) Id() int { return 0x01 }
