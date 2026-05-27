package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	setEnvDefault("EXTERNAL_API_URL", "https://mock.raysouz.studio")
	setEnvDefault("EXTERNAL_API_SERIAL", "b7af3a9e-6d1a-4b15-9837-3e0f0b47e5b4")

	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}

	var err error
	config, err = graphql.LoadConfig("examples/ecommerce-distributed/config/service.json")
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

	wrappedHandler = api.NewHandler(config.Pretty, graphql.CORSFromEnv())
	adapter = graphql.WrapHandler(wrappedHandler).ToALB()
}

func requestHandler(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	if req.Path == "/health" && req.HTTPMethod == http.MethodGet {
		return events.ALBTargetGroupResponse{StatusCode: http.StatusOK, Body: `{"status":"ok"}`}, nil
	}
	if req.Path != api.Config.Route || req.HTTPMethod != http.MethodPost {
		return events.ALBTargetGroupResponse{StatusCode: http.StatusNotFound, Body: `{"message":"not found"}`}, nil
	}
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	if _, ok := os.LookupEnv("ENVIRONMENT"); ok {
		lambda.Start(requestHandler)
		return
	}
	http.Handle(api.Config.Route, wrappedHandler)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8090"
	}
	fmt.Printf("Server running at http://localhost:%s%s\n", port, config.Route)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func setEnvDefault(key, value string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, value)
	}
}
