.PHONY: build deploy

build:
	sam build

deploy:
	sam deploy \
	--stack-name hello-world-sam \
	--region ap-northeast-1 \
	--confirm-changeset \
	--capabilities CAPABILITY_IAM \
	--no-disable-rollback \
	--parameter-overrides "HelloWorldFunctionAuth=NONE"

# dynamodbにテストデータを投入する
dynamodb-seed-test-data:
	aws dynamodb batch-write-item \
	--region ap-northeast-1 \
	--request-items '{
		"MyDynamoDBTable": [
			{
				"PutRequest": {
					"Item": {
						"id": {"S": "12345"},
						"Name": {"S": "Sample Name"},
						"Age": {"N": "30"}
					}
				}
			},
			{
				"PutRequest": {
					"Item": {
						"id": {"S": "67890"},
						"Name": {"S": "Another Name"},
						"Age": {"N": "25"}
					}
				}
			},
			{
				"PutRequest": {
					"Item": {
						"id": {"S": "67890sss"},
						"Name": {"S": "Another Name"},
						"Age": {"N": "250"}
					}
				}
			}
		]
	}'

# endpointを取得する
endpoints:
	sam list-endpoints --stack-name hello-world-sam

# 削除コマンド
delete:
	sam delete --stack-name <stack-name> --region <region>

# テストデータ取得
dynamodb-scan:
	aws dynamodb scan \
	--table-name MyDynamoDBTable \
	--region ap-northeast-1
# 以下参考
# aws dynamodb get-item \
#   --table-name MyDynamoDBTable \
#   --region ap-northeast-1 \
#   --key '{
#     "id": {"S": "12345"}
#   }'

# aws dynamodb query \
#   --table-name MyDynamoDBTable \
#   --region ap-northeast-1 \
#   --key-condition-expression "id = :id" \
#   --expression-attribute-values '{
#     ":id": {"S": "12345"}
#   }'
# ```
