# Go GraphQL Connector

Este projeto expõe uma fachada GraphQL configurável para consultar múltiplas origens de dados em paralelo. A ideia é adicionar ou remover integrações alterando configuração, sem mudar o código da aplicação.

## Configuração recomendada

Use um arquivo único de serviço:

```json
{
  "schema": "local:examples/config/schema.json",
  "connectors": "local:examples/config/connectors.json",
  "mock": "local:examples/config/mock.json",
  "route": "/graphql",
  "pretty": true,
  "graphiql": true,
  "allow_partial": false
}
```

Os campos `schema`, `connectors` e `mock` aceitam conteúdo inline ou paths no formato `local:`, `env:`, `ssm:` e `secrets:`.

## Conectores

Cada campo selecionado no GraphQL é resolvido por um conector com o mesmo nome:

```json
{
  "field": "convenio",
  "adapter": "redis",
  "adapterConfig": {
    "endpoint": "localhost:6379",
    "password": ""
  },
  "keyPattern": "CVN_{codigoConvenio}",
  "timeoutMs": 500,
  "retries": 1,
  "optional": false
}
```

Qualquer argumento da query pode ser usado em templates com `{nomeDoArgumento}`. Exemplo: `"/clientes/{clienteID}/contratos/{contratoID}"`.

## Recursos operacionais

- Execução paralela apenas dos campos pedidos na query.
- Timeout por conector via `timeoutMs`.
- Retry por conector via `retries`.
- Falhas parciais com `optional: true` por conector ou `allowPartial: true` no serviço.
- Mock por chave primária para testes locais e validação rápida.

## Execução local

```bash
go run ./examples
```

O endpoint padrão é `http://localhost:8080/graphql`.

## Validação

```bash
go test ./...
```
