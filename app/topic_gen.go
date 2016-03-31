package app
import (
	"io/ioutil"
	"github.com/shalinlk/wolf/models"
	"encoding/json"
)

var (
	topicGenPool map[string]*TopicGen
	stdQoS map[models.QoS][]models.QoS
)

func init() {
	topicGenPool = make(map[string]*TopicGen)
	stdQoS = map[models.QoS][]models.QoS{models.QOS_ZERO : []models.QoS{models.QOS_ZERO},
		models.QOS_ONE : []models.QoS{models.QOS_ONE},
		models.QOS_TWO : []models.QoS{models.QOS_TWO},
		models.QOS_LTE_ONE : []models.QoS{models.QOS_ZERO, models.QOS_ONE},
		models.QOS_LTE_TWO : []models.QoS{models.QOS_ZERO, models.QOS_ONE, models.QOS_TWO},
	}
}

type TopicGen struct {
	topicSource []topic
}

type topic struct {
	Topic string    `json:"topic"`
	QoS   *int      `json:"qos,omitempty"`
}

func NewTopicGen(path string) (*TopicGen, error) {
	tGen, found := topicGenPool[path]
	if found {
		return tGen, nil
	}
	topics, err := readTopics(path)
	if err != nil {
		return tGen, err
	}
	tGen.topicSource = topics
	topicGenPool[path] = tGen
	return tGen, nil
}

func (t *TopicGen)GetNewTopics(count int, patchQoS models.QoS) models.Topics {
	return t.newTopics(count, patchQoS)
}

//todo : dynamic topic creation from existing topic
func (t *TopicGen)newTopics(count int, patchQoS models.QoS) models.Topics {
	applicableQoS, found := stdQoS[patchQoS]
	if !found {
		applicableQoS = stdQoS[models.QOS_LTE_TWO]
	}
	applicableQoSLength := len(applicableQoS)
	topics := models.Topics{}
	topics = make(map[string]int)
	if count > len(t.topicSource) {
		count = len(t.topicSource)
	}
	for i := 0; i < count; i ++ {
		if t.topicSource[i].QoS == nil {
			topics[t.topicSource[i].Topic] = applicableQoS[i % applicableQoSLength]
		}else {
			topics[t.topicSource[i].Topic] = models.QoS(t.topicSource[i].QoS)
		}
	}
	return topics
}

func readTopics(path string) ([]topic, error) {
	var t []topic
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(fileBytes, &t)
	return t, err
}

