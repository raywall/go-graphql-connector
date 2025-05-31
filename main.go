package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/graphql-go/handler"
	"github.com/raywall/go-graphql-integrator/internal/graph"
	"github.com/raywall/go-graphql-integrator/internal/middleware"

	"github.com/raywall/cloud-easy-connector/pkg/aws"
	"github.com/raywall/cloud-easy-connector/pkg/local"
)

var (
	proxyHandler        *httpadapter.HandlerAdapterALB
	cloudContextFactory *aws.CloudContextFactory
	err                 error
)

func init() {
	// inicializa um contexto cloud
	cloudContextFactory, err = aws.NewCloudContextFactory("us-east-1", "http://localhost:4566")
	if err != nil {
		panic(err)
	}

	// inicializa um contexto de secrets manager
	schemaContext, err := cloudContextFactory.CreateContext(
		aws.SSMContext,
		map[string]interface{}{
			"parameter_name": local.New().GetEnvOrDefault("SSM_SCHEMA", "/graphql/dev/schema"),
		})
	if err != nil {
		panic(err)
	}

	connectorContext, err := cloudContextFactory.CreateContext(
		aws.SSMContext,
		map[string]interface{}{
			"parameter_name": local.New().GetEnvOrDefault("SSM_CONNECTORS", "/graphql/dev/connectors"),
		})
	if err != nil {
		panic(err)
	}

	// recupera o valor de um secrets manager
	schemaConfig, err := schemaContext.GetValue()
	if err != nil {
		panic(err)
	}

	connectorsConfig, err := connectorContext.GetValue()
	if err != nil {
		panic(err)
	}

	// Inicializar o resolver e o schema
	resolver, err := graph.NewResolver(connectorsConfig.(string))
	if err != nil {
		log.Fatalf("failed to create resolver: %v", err)
	}

	schema, err := graph.CreateSchema(resolver, schemaConfig.(string))
	if err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	// Configurar o handler GraphQL
	h := handler.New(
		&handler.Config{
			Schema:   schema,
			Pretty:   true,
			GraphiQL: true,
		})

	// Aplicar middleware chain
	wrappedHandler := middleware.Chain(
		h,
		// middleware.Logging,
		// middleware.Tracing,
	)

	// Adaptar o handler para Lambda
	proxyHandler = httpadapter.NewALB(wrappedHandler)
	http.Handle("/graphql", wrappedHandler)
}

func Handler(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	resp, err := proxyHandler.ProxyWithContext(ctx, req)
	if err != nil {
		return events.ALBTargetGroupResponse{}, err
	}

	return events.ALBTargetGroupResponse{
		StatusCode:        resp.StatusCode,
		Headers:           resp.Headers,
		MultiValueHeaders: resp.MultiValueHeaders,
		Body:              resp.Body,
		IsBase64Encoded:   resp.IsBase64Encoded,
	}, nil
}

func main() {
	// lambda.Start(Handler)

	fmt.Println("Server running at http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
