#!/bin/bash
set -e

# pip install localstack --break-system-packages
docker run --rm -it -d -p 4566:4566 -p 4510-4559:4510-4559 localstack/localstack --network local-kafka-broker_default --host localstack

aws --endpoint-url=http://localhost:4566 ssm put-parameter \
  --name "/graphql/dev/schema" \
  --type String \
  --value "$(cat config/schema.json)" \
  --region us-east-1

aws --endpoint-url=http://localhost:4566 ssm put-parameter \
  --name "/graphql/dev/connectors" \
  --type String \
  --value "$(cat config/connectors.json)" \
  --region us-east-1

aws --endpoint-url=http://localhost:4566 secretsmanager create-secret \
  --name /graphql/dev/schema \
  --secret-string "$(cat config/schema.json)" \
  --region us-east-1


aws --endpoint-url=http://localhost:4566 secretsmanager create-secret \
		--name "/graphql/dev/credentials" \
		--secret-string '{"client_id": "bbaca8bb-96d5-40cc-9e07-4b87b955119c", "client_secret": "2cf9a0b8-d7c2-4f19-b54f-92ccdc9996d5"}' \
		--region us-east-1

  