package shared

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var DBService dbService

type dbService struct {
	client    *dynamodb.Client
	tableName string
}

const databaseDefaultTimeout = 30 * time.Second

func (db *dbService) Init() {

	tableName := os.Getenv("CONTENT_MODERATION_TABLE_NAME")
	if tableName == "" {
		panic("table name is empty")
	}

	DBService.tableName = tableName

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("failed to load configs error, " + err.Error())
	}

	DBService.client = dynamodb.NewFromConfig(cfg)
}

func (db *dbService) PutItem(input interface{}) error {

	item, err := attributevalue.MarshalMapWithOptions(input, func(eo *attributevalue.EncoderOptions) {
		eo.TagKey = "json"
	})
	if err != nil {
		return err
	}

	putItemInput := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(db.tableName),
	}

	ctx, cancelFn := context.WithTimeout(context.TODO(), databaseDefaultTimeout)
	defer cancelFn()
	_, err = db.client.PutItem(ctx, putItemInput)
	return err
}

func (db *dbService) GetItem(id string, out interface{}) error {

	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		TableName: aws.String(db.tableName),
	}

	ctx, cancelFn := context.WithTimeout(context.TODO(), databaseDefaultTimeout)
	defer cancelFn()
	response, err := db.client.GetItem(ctx, getItemInput)
	if err != nil {
		return err
	}

	err = attributevalue.UnmarshalMap(response.Item, out)
	return err
}
