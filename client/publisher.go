package client
import (
	"time"
	"github.com/shalinlk/wolf/models"
"go/doc/testdata"
)

type AttackPlan struct {
	ignite chan struct{}
	publisher *Publisher
}

func newPlanePlanner() *AttackPlan {
	planeAttacker := &AttackPlan{}
	planeAttacker.ignite = make(chan struct{})
	return planeAttacker
}

func NewDurationBasedAttacker(duration time.Duration) *AttackPlan {
	attacker := newPlanePlanner()
	go func() {
		<-attacker.ignite
		//duration based attack goes here...
	}()
	return attacker
}

func NewMsgCountBasedAttacker(count int) *AttackPlan {
	attacker := newPlanePlanner()
	go func() {
		<-attacker.ignite
		//message count based attack goes here...
	}()
	return attacker
}

func (a *AttackPlan)LaunchAttack(startTime time.Time, publisher *Publisher) (error){
	if startTime.Before(time.Now()) {
		return models.ExpiredStartTimeError
	}
	if publisher == nil {
		return models.IvalidPublisherError
	}
	a.publisher = publisher
	<-time.After(startTime.Sub(time.Now()))
	a.ignite <- struct{}
	return nil
}

func (a *AttackPlan)attack()  {

}

type Publisher struct {

}


func (p *Publisher)RegisterTopics(map[string]int) {

}

func (p *Publisher)run() {
	
}