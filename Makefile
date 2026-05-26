.PHONY: test run-local run-dynamodb compose-up compose-down wait-dynamodb-config test-dynamodb

test:
	@go test -v ./...

run-local:
	@go run ./examples/local

compose-up:
	@docker compose up -d --remove-orphans --force-recreate
	@$(MAKE) wait-dynamodb-config

compose-down:
	@docker compose down --remove-orphans

wait-dynamodb-config:
	@echo "Aguardando configuracoes no DynamoDB LocalStack..."
	@set -e; \
	for i in $$(seq 1 60); do \
		if docker exec go-graphql-connector-localstack awslocal dynamodb get-item \
			--table-name graphql-config \
			--key '{"id":{"S":"connectors"}}' \
			--query 'Item.value.S' \
			--output text >/tmp/go-graphql-connector-connectors-check 2>/dev/null; then \
			if test "$$(cat /tmp/go-graphql-connector-connectors-check)" != "None"; then \
				echo "Configuracoes carregadas."; \
				exit 0; \
			fi; \
		fi; \
		sleep 1; \
	done; \
	echo "Timeout aguardando item connectors na tabela graphql-config."; \
	echo "Logs do LocalStack:"; \
	docker logs --tail 120 go-graphql-connector-localstack; \
	exit 1

run-dynamodb: compose-up
	@AWS_REGION=us-east-1 \
	AWS_DEFAULT_REGION=us-east-1 \
	AWS_ACCESS_KEY_ID=test \
	AWS_SECRET_ACCESS_KEY=test \
	AWS_ENDPOINT_URL=http://localhost:4566 \
	PORT=8081 \
	go run ./examples/dynamodb

test-dynamodb: compose-up
	@set -e; \
	AWS_REGION=us-east-1 \
	AWS_DEFAULT_REGION=us-east-1 \
	AWS_ACCESS_KEY_ID=test \
	AWS_SECRET_ACCESS_KEY=test \
	AWS_ENDPOINT_URL=http://localhost:4566 \
	PORT=8081 \
	go run ./examples/dynamodb > /tmp/go-graphql-connector-dynamodb.log 2>&1 & \
	pid=$$!; \
	trap 'kill $$pid >/dev/null 2>&1 || true' EXIT; \
	sleep 3; \
	curl -fsS -X POST http://localhost:8081/graphql \
		-H "Content-Type: application/json" \
		-d '{"query":"query { dataSources(codigoConvenio: 10341) { convenio { codigoConvenio nomeConvenio origem } limiteOperacional { percentualMaximoMargemConsignavel origem } } }"}' \
		| grep "Convênio de Crédito Servidores Estaduais" >/dev/null; \
	curl -fsS -X POST http://localhost:8081/graphql \
		-H "Content-Type: application/json" \
		-d '{"query":"query { dataSources(codigoConvenio: 10341) { convenio { origem } } }"}' \
		| grep '"origem": "elasticache' >/dev/null; \
	curl -fsS -X POST http://localhost:8081/graphql \
		-H "Content-Type: application/json" \
		-d '{"query":"query { dataSources(codigoConvenio: 10341) { limiteOperacional { origem } } }"}' \
		| grep '"origem": "api"' >/dev/null; \
	echo "examples/dynamodb ok"
