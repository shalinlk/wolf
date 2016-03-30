package models


type BasicPacket struct {
	Message string
	QoS     int
	Topic   string
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

type QoS int

var (
	QOS_ZERO QoS = 0
	QOS_ONE QoS = 1
	QOS_TWO QoS = 2
	QOS_LTE_ONE QoS = 3
	QOS_LTE_TWO QoS = 4
)