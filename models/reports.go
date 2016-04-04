package models

import "time"

type Stati interface {
	FinalReport() ([]byte, error)
}

type Report struct {
	Total    int            `json:"total"`
	Topics   map[string]int `json:"topics,omitempty"`
	Duration time.Duration  `json:"duration"`
}

type PublishReport struct {
	Publish Report `json:"publish,omitempty"`
}

func (p PublishReport) FinalReport() ([]byte, error) {
	return []byte{}, nil
}

type SubscribeReport struct {
	SubscribeReport Report `json:"subscribe,omitempty"`
}

func (s SubscribeReport) FinalReport() ([]byte, error) {
	return []byte{}, nil
}

type PubSubReport struct {
	PublishReport
	SubscribeReport
}

func (s PubSubReport) FinalReport() ([]byte, error) {
	return []byte{}, nil
}

type Statistics struct {
	SentTotal           int
	SentTopics          int
	SentDurationSec     int
	ReceivedTotal       int
	ReceivedTopics      int
	ReceivedDurationSec int
}
