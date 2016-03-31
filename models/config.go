package models


type Envs struct {
	TopicFile	string	`env:"topic_file;mandatory;/assets/gossip.json"`
}
