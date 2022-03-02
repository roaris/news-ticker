package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/line/line-bot-sdk-go/linebot"
)

func subscribe() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ddb := dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-northeast-1"))
	input := &dynamodb.ScanInput{TableName: aws.String("interests")}
	result, _ := ddb.Scan(input)

	for _, item := range result.Items {
		userID := *item["user_id"].S
		if _, err := bot.PushMessage(userID, linebot.NewTextMessage("定期送信")).Do(); err != nil {
			log.Print(err)
		}
	}
}

func main() {
	lambda.Start(subscribe)
}
