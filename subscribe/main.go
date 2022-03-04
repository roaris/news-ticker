package main

import (
	"fmt"
	"log"
	"os"

	"github.com/roaris/news-ticker/flex"

	"github.com/roaris/news-ticker/newsapi"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/line/line-bot-sdk-go/linebot"
)

var categoryEnToJa = map[string]string{
	"business":      "ビジネス",
	"entertainment": "エンタメ",
	"general":       "時事",
	"health":        "健康",
	"science":       "科学",
	"sports":        "スポーツ",
	"technology":    "テクノロジー",
}

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
				var bubbles []*linebot.BubbleContainer
				for _, article := range articlesWrapper.Articles {
					var bubble = flex.NewBubbleContainer(article)
					bubbles = append(bubbles, &bubble)
				}
				caroucel := flex.NewCaroucelContainer(bubbles)
				bot.PushMessage(userID, linebot.NewTextMessage(
					fmt.Sprintf("%sのニュースです", categoryEnToJa[categoryName])),
					linebot.NewFlexMessage(fmt.Sprintf("%sの記事です", categoryEnToJa[categoryName]), &caroucel)).Do()
			}
		}
	}
}

func main() {
	lambda.Start(subscribe)
}
