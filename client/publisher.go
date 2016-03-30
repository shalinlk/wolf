package client
import (
	"time"
	"github.com/shalinlk/wolf/models"
	"sync"
)

type AttackPlan struct {
	ignite         chan struct{}
	publisher      *Publisher
	doneSignal     sync.WaitGroup
	deliveryNozzle chan models.OutgoingPacket
}

func newPlanePlanner() *AttackPlan {
	planeAttacker := &AttackPlan{}
	planeAttacker.ignite = make(chan struct{})
	planeAttacker.deliveryNozzle = make(chan models.OutgoingPacket)
	return planeAttacker
}

func NewDurationBasedAttacker(duration time.Duration, conn *Conn) *AttackPlan {
	attacker := newPlanePlanner()
	conn.RegisterPublisher(attacker.deliveryNozzle)
	go func() {
		attacker.attack(func() {
			stopper := time.NewTimer(duration)
			for {
				select {
				case attacker.deliveryNozzle <- attacker.publisher.deliveryNozzle:
				case <-stopper.C:
					break
				}
			} })
	}()
	return attacker
}

func NewMsgCountBasedAttacker(count int, conn *Conn) *AttackPlan {
	attacker := newPlanePlanner()
	conn.RegisterPublisher(attacker.deliveryNozzle)
	go func() {
		attacker.attack(func() {
			for i := 0; i < count; i++ {
				attacker.deliveryNozzle <- attacker.publisher.deliveryNozzle
			}
		})
	}()
	return attacker
}

func (a *AttackPlan)LaunchAttack(startTime time.Time, publisher *Publisher) (error) {
	if startTime.Before(time.Now()) {
		return models.ExpiredStartTimeError
	}
	if publisher == nil {
		return models.InvalidPublisherError
	}
	a.publisher = publisher
	<-time.After(startTime.Sub(time.Now()))
	a.ignite <- struct{}
	a.doneSignal.Wait()
	return nil
}

func (a *AttackPlan)attack(attackController func()) {
	a.doneSignal.Add(1)
	defer a.publisher.done()
	defer a.doneSignal.Done()
	<-a.ignite
	attackController()
	close(a.deliveryNozzle)
}

type Publisher struct {
	deliveryNozzle chan models.OutgoingPacket
	topics         map[string]int
	doneSignal     chan struct{}
}

func (p *Publisher)NewPublisher(topics models.Topics) (*Publisher, error) {
	pub := &Publisher{}
	pub.doneSignal = make(chan struct{}, len(topics))
	return pub, nil
}

//NewGossipPublisher returns a publisher with random topics registered;
func (p *Publisher)NewGossipPublisher(topicNumber int) (*Publisher, error) {
	pub := &Publisher{}
	pub.doneSignal = make(chan struct{}, topicNumber)
	return pub, nil
}

func (p *Publisher)run() {
	for topic, qos := range p.topics {
		go func(topic string, qos int) {
			for {
				msg := models.OutgoingPacket{}
				msg.Message = "blaw. blaww.. blawww... balwwww.... balwwwww....."
				msg.QoS = qos
				msg.Topic = topic
				select {
				case <-p.doneSignal:
					break
				default:
					go func(msg models.OutgoingPacket) {
						p.deliveryNozzle <- msg
					}(msg)
				}
			}
		}(topic, qos)
	}
}

func (p *Publisher)done() {
	for i := 0; i < len(p.topics); i++ {
		p.doneSignal <- struct{}
	}
}