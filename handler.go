package graphql

import (
	hdl "github.com/graphql-go/handler"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type Handler http.Handler

func (g *GraphQL) NewHandler(pretty bool, middlewares ...Middleware) http.Handler {
	// Configurar o handler GraphQL
	h := &hdl.New(
		&hdl.Config{
			Schema:   g.Schema,
			Pretty:   Pretty,
			GraphiQL: true,
		})

	// Aplicar middleware chain
	return middleware.Chain(
		h,
		middlewares,
	)
}

func (h *Handler) ToAmazonALB() *httpadapter.HandlerAdapterALB {
	return httpadapter.NewALB(http.Handler(*h))
} 

func (h *Handler) ToAmazonAPIGateway() *httpadapter.HandlerAdapter {
	return httpadapter.New(http.Handler(*h))
} 

func (h *Handler) ToAmazonAPIGatewayV2() *httpadapter.HandlerAdapterV2 {
	return httpadapter.NewV2(http.Handler(*h))
} 