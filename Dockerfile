FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git && update-ca-certificates

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graphql-local ./examples/local
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graphql-pedidos ./examples/pedidos
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graphql-ecommerce-distributed ./examples/ecommerce-distributed
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graphql-dynamodb ./examples/dynamodb

# Corporate CA notes:
# - Alpine trusts certificates installed under /usr/local/share/ca-certificates.
# - To enable an internal CA, copy a .crt file to that directory and run
#   update-ca-certificates.
# - For AWS SDK/CLI calls that need a custom bundle, set:
#   ENV AWS_CA_BUNDLE=/usr/local/share/ca-certificates/internal-ca.crt
# - Example:
#   COPY certs/internal-ca.crt /usr/local/share/ca-certificates/internal-ca.crt
#   RUN update-ca-certificates
FROM alpine:3.22

WORKDIR /opt/go-graphql-connector

RUN apk add --no-cache ca-certificates curl && update-ca-certificates

ENV PORT=8090
ENV GRAPHQL_EXAMPLE=local

COPY --from=build /out/graphql-local /usr/local/bin/graphql-local
COPY --from=build /out/graphql-pedidos /usr/local/bin/graphql-pedidos
COPY --from=build /out/graphql-ecommerce-distributed /usr/local/bin/graphql-ecommerce-distributed
COPY --from=build /out/graphql-dynamodb /usr/local/bin/graphql-dynamodb
COPY examples ./examples

EXPOSE 8090

CMD ["/usr/local/bin/graphql-local"]
