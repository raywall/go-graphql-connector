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
	err            error
	api            *graphql.GraphQL
	config         *graphql.Config
)

func init() {
	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}

	config, err = graphql.LoadConfig("examples/local/config/service.json")
	if err != nil {
		config, err = graphql.LoadConfig("config/service.json")
	}
	if err != nil {
		panic(err)
	}

	api, err = graphql.New(config, resources, "us-east-1", "http://localhost:4566")
	if err != nil {
		panic(err)
	}

	// Configurar o handler GraphQL
	wrappedHandler = api.NewHandler(config.Pretty)

	// Adaptar o handler para Lambda
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

	} else if path != api.Config.Route || method != http.MethodPost {
		return events.ALBTargetGroupResponse{
			StatusCode: 404,
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
	} else {
		http.Handle(api.Config.Route, wrappedHandler)
		fmt.Printf("Server running at http://localhost:8080%s\n", config.Route)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
