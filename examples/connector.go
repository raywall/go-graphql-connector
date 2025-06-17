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
	"github.com/graphql-go/handler"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/go-graphql-connector/graphql"
	"github.com/raywall/go-graphql-connector/internal/middleware"
)

var (
	adapter        *httpadapter.HandlerAdapterALB
	wrappedHandler http.Handler
	err            error
	config         *graphql.Config
	api            *graphql.GraphQL
)

func init() {
	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}

	config = &graphql.Config{
		Schema:     "ssm:/graphql/dev/schema",
		Connectors: "ssm:/graphql/dev/connectors:false",
		Route:      "/graphql",
		Authorization: graphql.Authorization{
			RequireTokenSTS: true,
			TokenService: graphql.TokenService{
				TokenAuthorizationURL: "https://sts.teste.net/api/oauth/token",
				Credentials: graphql.Credentials{
					ClientID:     "env:STS_CLIENT_ID",
					ClientSecret: "secrets:/graphql/dev/credentials:json",
				},
			},
		},
	}

	api, err = graphql.New(config, resources, "us-east-1", "http://localhost:4566")
	if err != nil {
		panic(err)
	}

	// Configurar o handler GraphQL
	h := handler.New(
		&handler.Config{
			Schema:   api.Schema,
			Pretty:   true,
			GraphiQL: true,
		})

	// Aplicar middleware chain
	wrappedHandler = middleware.Chain(
		h,
		// middleware.Logging,
		// middleware.Tracing,
	)

	// Adaptar o handler para Lambda
	adapter = httpadapter.NewALB(wrappedHandler)
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

	} else if path != config.Route && method != http.MethodPost {
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
		http.Handle(config.Route, wrappedHandler)
		fmt.Println("Server running at http://localhost:8080/graphql")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
