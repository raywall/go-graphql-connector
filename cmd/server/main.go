package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/raywall/go-graphql-integrator/internal/graph"
	"github.com/raywall/go-graphql-integrator/internal/middleware"
)

var (
	schemaConfig     string = "./cmd/server/config/schema.json"
	connectorsConfig string = "./cmd/server/config/connectors.json"
)

func main() {
	resolver, err := graph.NewResolver(connectorsConfig)
	if err != nil {
		panic(err)
	}

	schema, err := graph.CreateSchema(resolver, schemaConfig)
	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Wrap handler with middleware chain
	wrappedHandler := middleware.Chain(
		h,
		// middleware.Logging,
		// middleware.Tracing,
	)

	http.Handle("/graphql", wrappedHandler)
	fmt.Println("Server running at http://localhost:8080/graphql")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
