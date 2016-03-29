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

type Topics map[string]int

func (t *Topics)AddTopic(topic string, qos int) {
	t[topic] = qos
}