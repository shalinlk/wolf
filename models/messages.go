package models


type BasicPacket struct {
	Message string
	QoS int
	Topic string
}

//OutgoingPacket defines the structure of atomic
//message to be send
type OutgoingPacket struct {
	BasicPacket
}

//IncomingPacket defines the structure of atomic
//messages received on the connection
type IncomingPacket struct {
	BasicPacket
}
