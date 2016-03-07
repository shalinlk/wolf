package client

import (
	"time"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"errors"
	"sync"
	"github.com/shalinlk/mqtt_paho_client_eval/models"
)

//todo : topic's QoS has to be considered
const (
	ATTACK_TYPE_NO_PLAN = 0
	LISTEN_TYPE_NO_PLAN = 0
	DEFAULT_QOS = 0
	ATTACK_TYPE_DURATION = 1
	LISTEN_TYPE_EXPLICIT = 1
	ATTACK_TYPE_MESSAGE_NUMBER = 2
	DEFAULT_TIME_OUT_SEC = 20
)

type Client struct {
	con                                  *MQTT.Client
	SentTopics, ReceivedTopics           map[string]int
	topicsToSend, TopicsToListen         map[string]int
	ClientId, UserId, Password           string
	StartingTime, ListenTill, AttackTill time.Time
	attackType, listenType               int
	listenTimeOutSec                     int
	receiveCountLock, sendCountLock      *sync.Mutex
	receiveSync, sendSync                chan string
	attackTunnel                         chan dataPack
	receiveTunnel                        chan string
	listenStopChan                       chan bool
	activeListener, activeAttacker       bool
	windUp                               sync.WaitGroup
}

//todo : choose the buffer sizes based up on a reason;
func NewClient(host, clientId, userId, password string) (*Client) {
	nC := &Client{UserId : userId, Password : password, ClientId : clientId,
		attackType : ATTACK_TYPE_NO_PLAN, listenType: LISTEN_TYPE_NO_PLAN,
		listenTimeOutSec : DEFAULT_TIME_OUT_SEC}
	nC.receiveSync = make(chan string, 10000)
	nC.sendSync = make(chan string, 10000)
	nC.attackTunnel = make(chan dataPack, 10000)
	nC.receiveTunnel = make(chan string, 10000)
	nC.listenStopChan = make(chan bool)
	opts := MQTT.NewClientOptions().AddBroker(host).SetClientID(clientId).SetUsername(userId).SetPassword(password)
	opts.SetDefaultPublishHandler(mqttListener(nC.receiveTunnel, nC.listenStopChan))

	c := MQTT.NewClient(opts)
	nC.con = c
	return nC
}

func mqttListener(recChan chan string, stopperChan chan bool) (func(*MQTT.Client, MQTT.Message)) {
	return func(client *MQTT.Client, msg MQTT.Message) {
		//	fmt.Printf("\rMSG: %s", msg.Payload())
		//	fmt.Printf("TOPIC: %s\n", msg.Topic())
		select {
		case <-stopperChan:
			close(recChan)
			return
		default:
			recChan <- msg.Topic()
		}
	}
}
func (c *Client)Connect() (error) {
	if token := c.con.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client)SetStartTime(startTime time.Time) (error) {
	if !time.Now().Before(startTime) {
		return errors.New("starting time already passed")
	}
	c.StartingTime = startTime
	return nil
}

func (c *Client)SetAttackDuration(attackDurationSeconds int) (error) {
	if c.StartingTime.IsZero() {
		return errors.New("starting time is not set")
	}
	c.AttackTill = c.StartingTime.Add(time.Second * time.Duration(attackDurationSeconds))
	c.attackType = ATTACK_TYPE_DURATION
	return nil
}

func (c *Client)SetListenDuration(listenDurationSeconds int) (error) {
	if c.StartingTime.IsZero() {
		return errors.New("starting time is not set")
	}
	c.ListenTill = c.StartingTime.Add(time.Second * time.Duration(listenDurationSeconds))
	c.listenType = LISTEN_TYPE_EXPLICIT
	return nil
}

func (c *Client)SetListenTimeOut(listenTimeOutSeconds int) {
	c.listenTimeOutSec = listenTimeOutSeconds
}

func (c *Client)RegisterTopicsToAttack(topics map[string]int) {
	c.topicsToSend = topics
	c.activeAttacker = true
}

func (c *Client)RegisterTopicsToListen(topics []string) {
	c.TopicsToListen = make(map[string]int)
	c.activeListener = true
	for _, topic := range topics {
		c.TopicsToListen[topic] = 0
	}
}

func (c *Client)ListenOnAttackingTopics() (error) {
	if !c.activeAttacker {
		return errors.New("no attacking topics registered")
	}
	//clean any previous trails
	c.TopicsToListen = make(map[string]int, len(c.topicsToSend))
	for topic, _ := range c.topicsToSend {
		c.TopicsToListen[topic] = 0
	}
	return nil
}

func (c *Client)Run() (error) {
	//raise the attacker
	if c.activeAttacker {
		go c.attacker()
		//select the attacker type
		switch c.attackType {
		case ATTACK_TYPE_DURATION:
			go c.attackForDuration()
		case ATTACK_TYPE_MESSAGE_NUMBER:
			go c.attackForMessageNumber()
		case ATTACK_TYPE_NO_PLAN:
			return errors.New("No attack plan..")
		default:
		}
	}
	//raise the listener
	if c.activeListener {
		c.listener()
	}
	return nil
}

func (c *Client)attacker() {
	for msg := range c.attackTunnel {
		c.con.Publish(msg.topic, byte(msg.qos), false, msg.payload)
		go func() {
			c.sendSync <- msg.topic
		}()
	}
}

func (c *Client)listener() {
	c.windUp.Add(1)
	defer func() {
		c.ListenTill = time.Now()
		c.windUp.Done()
	}()
	var stopper *time.Timer
	if c.listenType == LISTEN_TYPE_NO_PLAN {
		stopper = time.NewTimer(time.Second * time.Duration(c.listenTimeOutSec))
	}else if c.listenType == LISTEN_TYPE_EXPLICIT {
		stopper = time.NewTimer(c.ListenTill.Sub(c.StartingTime))
	}
	<-time.After(time.Now().Sub(c.StartingTime))
	for {
		select {
		case <-stopper.C:
			c.listenStopChan <- true
			return
		case topic := <-c.receiveTunnel:
			c.receiveSync <- topic
			if c.listenType == LISTEN_TYPE_NO_PLAN {
				//todo : test reset operation of shared variable
				stopper.Reset(time.Duration(time.Second * time.Duration(c.listenTimeOutSec)))
			}
		}
	}
}

func (c *Client)countReceived() {
	for topic := range c.receiveSync {
		c.receiveCountLock.Lock()
		if _, Ok := c.ReceivedTopics[topic]; Ok {
			c.ReceivedTopics[topic]++
		}else if _, Ok := c.TopicsToListen[topic]; Ok {
			c.ReceivedTopics[topic] = 1
		}
		c.receiveCountLock.Unlock()
	}
}

func (c *Client)countSent() {
	for topic := range c.sendSync {
		c.sendCountLock.Lock()
		if _, Ok := c.SentTopics[topic]; Ok {
			c.SentTopics[topic]++
		}else {
			c.SentTopics[topic] = 1
		}
		c.sendCountLock.Unlock()
	}
}

func (c *Client)attackForDuration() {
	c.windUp.Add(1)
	defer func() {
		close(c.attackTunnel)
		close(c.sendSync)
		c.windUp.Done()
	}()
	tillTime := time.NewTimer(c.AttackTill.Sub(c.StartingTime))
	waitTill := sync.WaitGroup{}
	for topic, _ := range c.topicsToSend {
		waitTill.Add(1)
		go func(topic string) {
			defer waitTill.Done()
			<-time.After(c.StartingTime.Sub(time.Now()))
			i := 0
			for {
				select {
				case <-tillTime.C:
					return
				default:
					c.attackTunnel <- dataPack{qos : DEFAULT_QOS, topic : topic, payload:i}
					go func(topic string) {
						c.sendSync <- topic
					}(topic)
				}
				i++
			}
		}(topic)
	}
	waitTill.Wait()
}

func (c *Client)attackForMessageNumber() {
	c.windUp.Add(1)
	defer func() {
		close(c.attackTunnel)
		close(c.sendSync)
		c.windUp.Done()
	}()
	waitTill := sync.WaitGroup{}
	for topic, count := range c.topicsToSend {
		waitTill.Add(1)
		go func(topic string, count int) {
			defer waitTill.Done()
			<-time.After(c.StartingTime.Sub(time.Now()))
			for i := 0; i < count; i++ {
				c.attackTunnel <- dataPack{qos: DEFAULT_QOS, topic: topic, payload:i}
				go func(topic string) {
					c.sendSync <- topic
				}(topic)
			}
		}(topic, count)
	}
	waitTill.Wait()
	c.AttackTill = time.Now()
}

func (c *Client)Statistics() (models.Statistics, error) {
	c.windUp.Wait()
	result := models.Statistics{}
	if c.activeAttacker {
		for _, count := range c.SentTopics {
			result.SentTotal += count
		}
		result.SentTopics = c.SentTopics
		result.SentDurationSec = int(c.AttackTill.Unix() - c.StartingTime.Unix())
	}
	if c.activeListener {
		for _, count := range c.ReceivedTopics {
			result.ReceivedTotal += count
		}
		result.ReceivedTopics = c.ReceivedTopics
		result.ReceivedDurationSec = int(c.ListenTill.Unix() - c.StartingTime.Unix())
	}
	return result, nil
}

type dataPack struct {
	qos     int
	topic   string
	;payload interface{}
}