package main

import (
	"log"
	"os"

	"github.com/roaris/news-ticker/newsapi"

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

		for _, category := range item["categories"].L {
			categoryName := *category.S
			articlesWrapper, err := newsapi.RequestArticles(categoryName)

			if err != nil {
				bot.PushMessage(userID, linebot.NewTextMessage("ニュースの取得に失敗しました...")).Do()
			} else {
				for _, article := range articlesWrapper.Articles {
					bot.PushMessage(userID, linebot.NewTextMessage(article.Title)).Do()
				}
			}
		}
	}
}

func main() {
	lambda.Start(subscribe)
}
