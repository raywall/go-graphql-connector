package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/graphql-go/handler"
	"github.com/raywall/go-graphql-integrator/internal/graph"
	"github.com/raywall/go-graphql-integrator/internal/middleware"

	"github.com/raywall/cloud-easy-connector/pkg/auth"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-easy-connector/pkg/local"
)

var (
	adapter        *httpadapter.HandlerAdapterALB
	cloudContext   cloud.CloudContext
	wrappedHandler http.Handler
	err            error
	route          string = "/graphql"
)

func init() {
	// inicializa um contexto cloud
	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}
	cloudContext, err = cloud.NewAwsCloudContext("us-east-1", "http://localhost:4566", resources)
	if err != nil {
		panic(err)
	}

	// recupera o valor dos parametros
	sch, err := cloudContext.GetParameterValue(local.New().GetEnvOrDefault("SSM_SCHEMA", "/graphql/dev/schema"), false)
	if err != nil {
		panic(err)
	}

	con, err := cloudContext.GetParameterValue(local.New().GetEnvOrDefault("SSM_SCHEMA", "/graphql/dev/connectors"), false)
	if err != nil {
		panic(err)
	}

	// recupera o valor do secret
	authRequest := auth.AuthRequest{}
	jsonSecretsValue, err := cloudContext.GetSecretValue(local.New().GetEnvOrDefault("SECRET_CREDENTIALS", "/graphql/dev/credentials"), cloud.JSONSecret)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonSecretsValue.([]byte), &authRequest)
	if err != nil {
		panic(err)
	}

	// inicializa um token client auto gerenciado
	cloudContext.NewAutoManagedToken(
		local.New().GetEnvOrDefault("AUTH_BASE_URL", "https://sts.teste.net/api/oauth/token"),
		authRequest.ClientID,
		authRequest.ClientSecret,
		false)

	// if err = cloudContext.GetAutoManagedToken().Start(); err != nil {
	// 	panic(err)
	// }

	// Inicializar o resolver e o schema
	resolver, err := graph.NewResolver(con.(string))
	if err != nil {
		log.Fatalf("failed to create resolver: \n\t%v", err)
	}
	if err := resolver.AddCloudContext(cloudContext); err != nil {
		log.Fatalf("failed to add a cloud context: \n\t%v", err)
	}

	schema, err := graph.CreateSchema(resolver, sch.(string))
	if err != nil {
		log.Fatalf("failed to create schema: \n\t%v", err)
	}

	// Configurar o handler GraphQL
	h := handler.New(
		&handler.Config{
			Schema:   schema,
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

	} else if path != route && method != http.MethodPost {
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
		http.Handle(route, wrappedHandler)
		fmt.Println("Server running at http://localhost:8080/graphql")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
