package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/roaris/news-ticker/models"
)

var categories = map[string]struct{}{
	"ビジネス":   struct{}{},
	"エンタメ":   struct{}{},
	"時事":     struct{}{},
	"健康":     struct{}{},
	"科学":     struct{}{},
	"スポーツ":   struct{}{},
	"テクノロジー": struct{}{},
}

func concatCategories(categories []*dynamodb.AttributeValue) string {
	sentence := ""
	for i, category := range categories {
		sentence += *category.S
		if i < len(categories)-1 {
			sentence += "、"
		}
	}
	return sentence
}

func replyMessage(bot *linebot.Client, replyToken string, message string) {
	if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do(); err != nil {
		log.Print(err)
	}
}

func handleAdd(bot *linebot.Client, event *linebot.Event, ddb *dynamodb.DynamoDB, addCategory string) {
	if _, ok := categories[addCategory]; !ok {
		replyMessage(bot, event.ReplyToken, "カテゴリはビジネス、エンタメ、時事、健康、科学、スポーツ、テクノロジーから選ぶことができます")
		return
	}
	categories, err := models.GetCategories(ddb, event.Source.UserID)
	if err != nil {
		replyMessage(bot, event.ReplyToken, "カテゴリの追加に失敗しました...")
		return
	}
	for _, category := range categories {
		if *category.S == addCategory {
			replyMessage(bot, event.ReplyToken, fmt.Sprintf("既にそのカテゴリは追加済みです\n現在のカテゴリは%sです", concatCategories(categories)))
			return
		}
	}
	if err := models.AddCategory(ddb, event.Source.UserID, addCategory); err != nil {
		replyMessage(bot, event.ReplyToken, "カテゴリの追加に失敗しました...")
		return
	}
	newCategories := append(categories, &dynamodb.AttributeValue{S: aws.String(addCategory)})
	replyMessage(bot, event.ReplyToken, fmt.Sprintf("カテゴリを追加しました\n現在のカテゴリは%sです", concatCategories(newCategories)))
}

func handleRemove(bot *linebot.Client, event *linebot.Event, ddb *dynamodb.DynamoDB, removeCategory string) {
	if _, ok := categories[removeCategory]; !ok {
		replyMessage(bot, event.ReplyToken, "カテゴリはビジネス、エンタメ、時事、健康、科学、スポーツ、テクノロジーから選ぶことができます")
		return
	}
	categories, err := models.GetCategories(ddb, event.Source.UserID)
	if err != nil {
		replyMessage(bot, event.ReplyToken, "カテゴリの削除に失敗しました...")
		return
	}
	for i, category := range categories {
		if *category.S == removeCategory {
			if err := models.RemoveCategory(ddb, event.Source.UserID, i); err != nil {
				replyMessage(bot, event.ReplyToken, "カテゴリの削除に失敗しました...")
			} else {
				newCategories := append(categories[:i], categories[i+1:]...)
				var message string
				if len(newCategories) > 0 {
					message = fmt.Sprintf("カテゴリを削除しました\n現在のカテゴリは%sです", concatCategories(newCategories))
				} else {
					message = "カテゴリを削除しました\nニュースの送信を停止します"
				}
				replyMessage(bot, event.ReplyToken, message)
			}
			return
		}
	}
	var message string
	if len(categories) > 0 {
		message = fmt.Sprintf("そのカテゴリは追加されていません\n現在のカテゴリは%sです", concatCategories(categories))
	} else {
		message = "そのカテゴリは追加されていません\n現在登録されているカテゴリはありません"
	}
	replyMessage(bot, event.ReplyToken, message)
}

func validateSignature(channelSecret, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))

	_, err = hash.Write(body)
	if err != nil {
		return false
	}

	return hmac.Equal(decoded, hash.Sum(nil))
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	body := []byte(request.Body)

	// リクエストがLINEから送られたものであることを検証する
	if !validateSignature(os.Getenv("CHANNEL_SECRET"), request.Headers["x-line-signature"], body) {
		return events.APIGatewayProxyResponse{StatusCode: 400}, linebot.ErrInvalidSignature
	}

	payload := &struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err := json.Unmarshal(body, payload); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	ddb := dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-northeast-1"))

	for _, event := range payload.Events {
		if event.Type == linebot.EventTypeFollow { // 友達追加時とブロック解除時
			// あいさつメッセージは管理画面で設定
			// デフォルトの興味は時事のみ
			userID := event.Source.UserID
			input := &dynamodb.PutItemInput{
				Item: map[string]*dynamodb.AttributeValue{
					"user_id":    {S: aws.String(userID)},
					"categories": {L: []*dynamodb.AttributeValue{&dynamodb.AttributeValue{S: aws.String("時事")}}},
				},
				TableName: aws.String("interests"),
			}
			if _, err = ddb.PutItem(input); err != nil {
				log.Print(err)
			}
		} else if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				text := message.Text
				switch text[0] {
				case '+':
					addCategory := text[1:]
					handleAdd(bot, event, ddb, addCategory)
				case '-':
					removeCategory := text[1:]
					handleRemove(bot, event, ddb, removeCategory)
				default:
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
