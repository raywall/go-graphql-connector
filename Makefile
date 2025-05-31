# Create Docker network and run LocalStack container
localstack:
	docker run --rm -d \
		-p 4566:4566 \
		-p 4510-4559:4510-4559 \
		--network local-kafka-broker_default \
		--hostname localstack \
		--name localstack \
		localstack/localstack

	@if [ ! -f config/schema.json ]; then echo "Error: config/schema.json not found"; exit 1; fi
	@if [ ! -f config/connectors.json ]; then echo "Error: config/connectors.json not found"; exit 1; fi
	aws --endpoint-url=http://localhost:4566 ssm put-parameter \
		--name "/graphql/dev/schema" \
		--type String \
		--value "$$(cat config/schema.json)" \
		--region us-east-1 \
		--overwrite
	aws --endpoint-url=http://localhost:4566 ssm put-parameter \
		--name "/graphql/dev/connectors" \
		--type String \
		--value "$$(cat config/connectors.json)" \
		--region us-east-1 \
		--overwrite

# Run locally
local:
	go run .

# Build and run SAM locally with LocalStack integration
run:	
	GOOS=linux GOARCH=amd64 go build -o main .
	sam local start-api \
		--region us-east-1 \
		--docker-network local-kafka-broker_default \
		--warm-containers eager

# Clean up
clean:
	-docker stop localstack
	-rm main

.PHONY: localstack local run clean