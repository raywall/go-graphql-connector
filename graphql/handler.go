package graphql

import (
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	hdl "github.com/graphql-go/handler"
	"github.com/raywall/go-graphql-connector/internal/middleware"
)

type Middleware = middleware.Middleware

type Handler struct {
	http.Handler
}

func (g *GraphQL) NewHandler(pretty bool, middlewares ...Middleware) http.Handler {
	h := hdl.New(&hdl.Config{
		Schema:   g.Schema,
		Pretty:   pretty,
		GraphiQL: g.Config.GraphiQL,
	})

	chain := append([]Middleware{middleware.ElapsedTime}, middlewares...)
	return middleware.Chain(h, chain...)
}

func WrapHandler(h http.Handler) Handler {
	return Handler{Handler: h}
}

func (h Handler) ToALB() *httpadapter.HandlerAdapterALB {
	return httpadapter.NewALB(h.Handler)
}

func (h Handler) ToAPIGateway() *httpadapter.HandlerAdapter {
	return httpadapter.New(h.Handler)
}

func (h Handler) ToAPIGatewayV2() *httpadapter.HandlerAdapterV2 {
	return httpadapter.NewV2(h.Handler)
}

func (h Handler) ToAmazonALB() *httpadapter.HandlerAdapterALB {
	return h.ToALB()
}

func (h Handler) ToAmazonAPIGateway() *httpadapter.HandlerAdapter {
	return h.ToAPIGateway()
}

func (h Handler) ToAmazonAPIGatewayV2() *httpadapter.HandlerAdapterV2 {
	return h.ToAPIGatewayV2()
}
