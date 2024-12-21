package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	var cfg aws.Config
	var err error

	// 環境変数でローカル実行かどうかを判定
	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		// ローカルDynamoDB Local用の設定
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{
						URL: "http://host.docker.internal:8000", // Dockerで起動したDynamoDB Localを指す
						// URL:           "http://localhost:8000", // Dockerで起動したDynamoDB Localを指す
						SigningRegion: "ap-northeast-1",
					}, nil
				}
				return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
			})),
		)
		log.Println("Using DynamoDB Local")
	} else {
		// クラウドDynamoDB用の設定
		cfg, err = config.LoadDefaultConfig(context.TODO())
		log.Println("Using DynamoDB in AWS Cloud")
	}

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)

	// DynamoDBテーブル名を環境変数から取得
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME environment variable is not set")
	}

	log.Printf("DynamoDB table name: %s", tableName)
}

// sam deploy したあとにどのようにクラウドリソースを更新するのか？

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// データをDynamoDBに保存
	err := putItemToDynamoDB("1", "Alice", "20")
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

	// クライアントのIPアドレスを取得
	sourceIP := request.RequestContext.Identity.SourceIP
	if sourceIP == "" {
		sourceIP = "unknown"
	}
	log.Printf("Source IP: %s", sourceIP)

	// レスポンス作成
	responseBody := fmt.Sprintf("Hello, %s!\nStored data: %v\n", sourceIP, items)
	return events.APIGatewayProxyResponse{
		Body:       responseBody,
		StatusCode: 200,
	}, nil
}

func putItemToDynamoDB(id, name, age string) error {
	tableName := "MyDynamoDBTable"
	_, err := dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: id},
			"Name": &types.AttributeValueMemberS{Value: name},
			"Age":  &types.AttributeValueMemberN{Value: age},
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
