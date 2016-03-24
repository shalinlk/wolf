package client

import (
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/shalinlk/wolf/models"
	"fmt"
)

var connPool map[string]*Conn

//Conn defines the underlying MQTT connection
//which is used for sending and receiving message
type Conn struct {
	con             *MQTT.Client
	broadcastTunnel chan models.OutgoingPacket
	receiverTunnel  chan models.IncomingPacket
	subscribers     map[string]chan models.IncomingPacket
}

func init() {
	connPool = make(map[string]*Conn)
}

func NewConnection(userId, clientId, password, host string) (*Conn, error) {
	signature := fmt.Sprintf("%s.%s.%s.%s", userId, clientId, password, host)
	conn, found := connPool[signature]
	if found {
		return conn, nil
	}
	//a pool miss
	opts := MQTT.NewClientOptions().AddBroker(host).SetClientID(clientId).SetUsername(userId).SetPassword(password)
	opts.SetDefaultPublishHandler(conn.packetObserver())
	conn.con = MQTT.NewClient(opts)
	connPool[signature] = conn
	conn.packetDistributor()
	return conn, nil
}

func (c *Conn) RegisterPublisher(publisher <- chan models.OutgoingPacket) {
	c.packetFlusher(publisher)
}

func (c *Conn) RegisterSubscriber() <- chan models.IncomingPacket {
	broadcastPipe := make(chan models.IncomingPacket)
	c.subscribers[len(c.subscribers)] = broadcastPipe
	return broadcastPipe
}

func (c *Conn) packetDistributor() {
	for packet := range c.receiverTunnel {
		for _, receiver := range c.subscribers {
			//we don't want to be blocked by sloths ;~)
			go func(pkt models.IncomingPacket) {
				receiver <- pkt
			}(packet)
		}
	}
}

func (c *Conn) packetFlusher(publisher chan models.OutgoingPacket) {
	go func(ch chan models.OutgoingPacket) {
		for packet := range ch {
			c.broadcastTunnel <- packet
		}
	}(publisher)
}

func (c *Conn) packetObserver() func(*MQTT.Client, MQTT.Message) {
	return func(client *MQTT.Client, msg MQTT.Message) {
		c.receiverTunnel <- mqttMsgToWolfMsg(msg)
	}
}

func mqttMsgToWolfMsg(msg MQTT.Message) models.IncomingPacket {
	inMsg := models.IncomingPacket{}
	inMsg.Message = string(msg.Payload())
	inMsg.Topic = msg.Topic()
	inMsg.QoS = int(msg.Qos())
	return inMsg
}