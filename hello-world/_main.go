package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dynamoClient *dynamodb.Client
	tableName    string
)

func init() {
	// DynamoDBクライアントの初期化
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)

	// DynamoDBテーブル名を環境変数から取得
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME environment variable is not set")
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sourceIP := request.RequestContext.Identity.SourceIP

	if sourceIP == "" {
		sourceIP = "unknown"
	}

	// データをDynamoDBに保存
	err := putItemToDynamoDB(sourceIP)
	if err != nil {
		log.Printf("Failed to put item to DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			Body:       "Failed to store data in DynamoDB\n",
			StatusCode: 500,
		}, nil
	}

	// DynamoDBからデータを取得
	items, err := getItemsFromDynamoDB()
	if err != nil {
		log.Printf("Failed to get items from DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			Body:       "Failed to retrieve data from DynamoDB\n",
			StatusCode: 500,
		}, nil
	}

	// レスポンス作成
	responseBody := fmt.Sprintf("Hello, %s!\nStored data: %v\n", sourceIP, items)
	return events.APIGatewayProxyResponse{
		Body:       responseBody,
		StatusCode: 200,
	}, nil
}

func putItemToDynamoDB(sourceIP string) error {
	_, err := dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item: map[string]types.AttributeValue{
			"source_ip": &types.AttributeValueMemberS{
				Value: sourceIP,
			},
		},
	})
	return err
}

func getItemsFromDynamoDB() ([]map[string]types.AttributeValue, error) {
	output, err := dynamoClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &tableName,
	})
	if err != nil {
		return nil, err
	}
	return output.Items, nil
}

func main() {
	lambda.Start(handler)
}
