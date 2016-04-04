package main

import (
	"fmt"
	"github.com/shalinlk/wolf/client"
	"github.com/HiFX/env_parser"
	"github.com/shalinlk/wolf/models"
"github.com/shalinlk/wolf/app"
	"time"
)

func main() {
	fmt.Println("Wolfs are not Bees..")
	ep := env_parser.NewEnvParser()
	envs := models.Envs{}
	err := ep.Map(&envs)
	if err != nil {
		fmt.Println("Invalid environment variables : ", err )
		return
	}
	conn, err := client.NewConnection(envs.UserId, envs.ClientId, envs.Password, envs.Host)
	if err != nil {
		fmt.Println("Error in obtaining connection : ", err)
		return
	}
	topicGen, err := app.NewTopicGen(envs.TopicFile)
	if err != nil {
		fmt.Println("Topic generator obtaining error : ", err)
		return
	}
	publisher, err := client.NewPublisher(topicGen.GetNewTopics(50, models.QOS_LTE_ONE))
	if err != nil {
		fmt.Println("Publisher obtaining error : ", err)
		return
	}
	attacker := client.NewMsgCountBasedAttacker(100000, conn)
	report, err := attacker.LaunchAttack(time.Now().Add(time.Duration(time.Second * 10)), publisher)
	if err != nil {
		fmt.Println("Attacking error : ", err)
		return
	}
	rep, err := report.FinalReport()
	if err != nil {
		fmt.Println("Report gathering error : ", err)
		return
	}
	fmt.Println("Final Report : ", string(rep))
}
