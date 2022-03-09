package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// 特定のユーザーのカテゴリを取得する
func GetCategories(ddb *dynamodb.DynamoDB, userID string) ([]*dynamodb.AttributeValue, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {S: aws.String(userID)},
		},
		TableName: aws.String("interests"),
	}
	result, err := ddb.GetItem(input)
	if err != nil {
		return nil, err
	}
	return result.Item["categories"].L, nil
}

// カテゴリを追加する
func AddCategory(ddb *dynamodb.DynamoDB, userID string, category string) error {
	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {S: aws.String(userID)},
		},
		ExpressionAttributeNames: map[string]*string{
			"#C": aws.String("categories"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":c": {L: []*dynamodb.AttributeValue{&dynamodb.AttributeValue{S: aws.String(category)}}},
		},
		UpdateExpression: aws.String("SET #C = list_append(#C, :c)"),
		TableName:        aws.String("interests"),
	}
	_, err := ddb.UpdateItem(input)
	return err
}

// カテゴリを削除する
func RemoveCategory(ddb *dynamodb.DynamoDB, userID string, index int) error {
	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {S: aws.String(userID)},
		},
		UpdateExpression: aws.String(fmt.Sprintf("REMOVE categories[%d]", index)),
		TableName:        aws.String("interests"),
	}
	_, err := ddb.UpdateItem(input)
	return err
}
