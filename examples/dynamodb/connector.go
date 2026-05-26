package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/go-graphql-connector/graphql"
)

var (
	adapter        *httpadapter.HandlerAdapterALB
	wrappedHandler http.Handler
	api            *graphql.GraphQL
	config         *graphql.Config
)

func init() {
	setEnvDefault("AWS_REGION", "us-east-1")
	setEnvDefault("AWS_ACCESS_KEY_ID", "test")
	setEnvDefault("AWS_SECRET_ACCESS_KEY", "test")
	setEnvDefault("AWS_ENDPOINT_URL", "http://localhost:4566")

	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}

	config = &graphql.Config{
		Schema:     "dynamodb:us-east-1:graphql-config:schema",
		Connectors: "dynamodb:us-east-1:graphql-config:connectors",
		Route:      "/graphql",
		Pretty:     true,
		GraphiQL:   true,
	}

	var err error
	api, err = graphql.New(config, resources, "us-east-1", os.Getenv("AWS_ENDPOINT_URL"))
	if err != nil {
		panic(err)
	}

	wrappedHandler = api.NewHandler(config.Pretty, graphql.CORSFromEnv())
	adapter = graphql.WrapHandler(wrappedHandler).ToALB()
}

func requestHandler(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	method := req.HTTPMethod
	path := req.Path

	if path == "/health" && method == http.MethodGet {
		return events.ALBTargetGroupResponse{
			StatusCode: http.StatusOK,
			Body:       `{"status": "ok"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	if path != api.Config.Route || method != http.MethodPost {
		return events.ALBTargetGroupResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"message": "rota não encontrada ou método não permitido"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	if _, ok := os.LookupEnv("ENVIRONMENT"); ok {
		lambda.Start(requestHandler)
		return
	}

	port := getEnvDefault("PORT", "8081")
	http.Handle(api.Config.Route, wrappedHandler)
	fmt.Printf("DynamoDB example running at http://localhost:%s%s\n", port, api.Config.Route)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func setEnvDefault(name, value string) {
	if os.Getenv(name) == "" {
		_ = os.Setenv(name, value)
	}
}

func getEnvDefault(name, value string) string {
	if current := os.Getenv(name); current != "" {
		return current
	}
	return value
}
