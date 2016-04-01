package models

type Envs struct {
	TopicFile string `env:"topic_file;mandatory;/assets/gossip.json"`
	UserId    string `env:"user_id;mandatory"`
	ClientId  string `env:"client_id;mandatory"`
	Password  string `env:"password;mandatory"`
	Host      string `env:"host;mandatory"`
}
