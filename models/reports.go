package models
import "time"

type Stati interface {
	Report() ([]byte, error)
}

type Report struct {
	Total    int    			`json:"total"`
	Topics   map[string]int 	`json:"topics,omitempty"`
	Duration time.Duration    	`json:"duration"`
}

type PublishReport struct {
	Publish Report    `json:"publish,omitempty"`
}

func (p PublishReport)Report() ([]byte, error) {
	return []byte, nil
}

type SubscribeReport struct {
	SubscribeReport Report    `json:"subscribe,omitempty"`
}

func (s SubscribeReport )Report() ([]byte, error) {
	return []byte, nil
}

type PubSubReport struct {
	PublishReport
	SubscribeReport
}

func (s PubSubReport) Report() ([]byte, error) {
	return []byte, nil
}

type Statistics struct {
	SentTotal           int
	SentTopics          int
	SentDurationSec     int
	ReceivedTotal       int
	ReceivedTopics      int
	ReceivedDurationSec int
}