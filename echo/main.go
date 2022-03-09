package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/line/line-bot-sdk-go/linebot"
)

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
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
