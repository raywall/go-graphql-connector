# Go GraphQL Connector - Documentacao

O Go GraphQL Connector cria uma API GraphQL dinamica a partir de configuracao. A aplicacao informa o schema, os conectores e as fontes de configuracao; o projeto monta o handler GraphQL e resolve cada campo solicitado buscando dados nas origens configuradas.

## Recursos Disponiveis

- Schema GraphQL dinamico via JSON.
- Conectores por campo GraphQL.
- Busca paralela dos campos solicitados na query.
- Suporte a Redis, REST, DynamoDB, S3 e RDS.
- Templates dinamicos em chaves e endpoints, como `CVN_{codigoConvenio}`.
- Configuracao via arquivo local, variavel de ambiente, AWS Systems Manager Parameter Store, AWS Secrets Manager, S3 ou DynamoDB.
- Mock configuravel para testes locais.
- Timeout, retry e falha opcional por conector.
- Adaptadores para HTTP local, AWS ALB, API Gateway v1 e API Gateway v2.
- Header `x-graphql-elapsed-time` em respostas GraphQL com o tempo de processamento em milissegundos.

## Exemplo Rapido

Execute o exemplo local:

```bash
make run-local
```

Endpoint:

```text
http://localhost:8080/graphql
```

Query de teste:

```graphql
query {
  dataSources(codigoConvenio: 10341) {
    convenio {
      codigoConvenio
      nomeConvenio
    }
    limiteOperacional {
      percentualMaximoMargemConsignavel
    }
    taxaFunding {
      dataFunding
    }
  }
}
```

Com `curl`:

```bash
curl -i -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { dataSources(codigoConvenio: 10341) { convenio { codigoConvenio nomeConvenio } limiteOperacional { percentualMaximoMargemConsignavel } taxaFunding { dataFunding } } }"}'
```

A resposta inclui o header:

```http
x-graphql-elapsed-time: 3
```

O valor representa milissegundos e retorna apenas o numero.

## Exemplo com Configuracao no DynamoDB, Dados no Redis e API REST

O projeto tambem possui um exemplo em `examples/dynamodb` que carrega `schema` e `connectors` a partir de uma tabela DynamoDB. Nesse exemplo, o DynamoDB e usado apenas como origem de configuracao; os dados consultados pela GraphQL API vêm de Redis e de uma API HTTP.

Suba a infraestrutura local:

```bash
make compose-up
```

O `docker-compose.yml` sobe:

- LocalStack com DynamoDB.
- Redis compativel com ElastiCache para o modo Community.
- Uma API HTTP local para simular uma integracao REST.

O script `scripts/localstack/init/01-dynamodb-config.sh` cria a tabela `graphql-config`, popula a chave `CVN_10341` no Redis e grava dois itens no DynamoDB:

- `id = schema`
- `id = connectors`

Cada item usa o atributo `value` contendo o JSON da respectiva configuracao.

Observacao: ElastiCache no LocalStack Community retorna `pro feature`. Por isso, o exemplo padrao usa um container Redis local como substituto compativel com ElastiCache. Se voce tiver LocalStack Pro, execute com:

```bash
USE_LOCALSTACK_ELASTICACHE=true LOCALSTACK_AUTH_TOKEN=seu_token make run-dynamodb
```

Nesse modo, o script tenta criar o cache cluster pelo servico ElastiCache do LocalStack.

Execute o exemplo:

```bash
make run-dynamodb
```

Execute o teste automatizado do exemplo:

```bash
make test-dynamodb
```

O exemplo usa estes paths:

```go
config = &graphql.Config{
    Schema:     "dynamodb:us-east-1:graphql-config:schema",
    Connectors: "dynamodb:us-east-1:graphql-config:connectors",
    Route:      "/graphql",
    Pretty:     true,
    GraphiQL:   true,
}
```

Query de teste:

```graphql
query {
  dataSources(codigoConvenio: 10341) {
    convenio {
      codigoConvenio
      nomeConvenio
      origem
    }
    limiteOperacional {
      percentualMaximoMargemConsignavel
      origem
    }
  }
}
```

Nesse retorno:

- `convenio` vem do Redis compativel com ElastiCache no modo Community, ou do ElastiCache do LocalStack quando `USE_LOCALSTACK_ELASTICACHE=true`.
- `limiteOperacional` vem da API REST.

Para encerrar o LocalStack:

```bash
make compose-down
```

## Configuracao Principal

O arquivo de configuracao principal pode ser carregado com `graphql.LoadConfig`.

Exemplo em `examples/config/service.json`:

```json
{
  "version": "1",
  "schema": "local:schema.json",
  "connectors": "local:connectors.json",
  "mock": "local:mock.json",
  "route": "/graphql",
  "pretty": true,
  "graphiql": true,
  "allow_partial": false
}
```

Campos:

- `schema`: conteudo inline do schema ou path para captura do schema.
- `connectors`: conteudo inline dos conectores ou path para captura dos conectores.
- `mock`: conteudo inline do mock ou path para captura do mock. Campo opcional.
- `route`: rota HTTP usada pela API. Padrao: `/graphql`.
- `pretty`: habilita resposta GraphQL formatada.
- `graphiql`: habilita interface GraphiQL no handler.
- `allow_partial`: permite que falhas em conectores retornem campos nulos sem falhar a consulta inteira.
- `authorization`: configuracao opcional para token STS usado por conectores.

## Fontes de Captura

Os campos `schema`, `connectors`, `mock`, `token_authorization_url`, `client_id` e `client_secret` aceitam valores inline ou paths.

### Conteudo Inline

Use o JSON diretamente no campo:

```json
{
  "schema": "{\"types\":[],\"query\":{\"name\":\"Query\",\"fields\":[]}}",
  "connectors": "{\"connectors\":[]}"
}
```

### Arquivo Local

Formato:

```text
local:caminho/do/arquivo.json
```

Exemplo:

```json
{
  "schema": "local:schema.json",
  "connectors": "local:connectors.json",
  "mock": "local:mock.json"
}
```

Quando usado dentro de um arquivo carregado por `graphql.LoadConfig`, paths relativos sao resolvidos a partir do diretorio do arquivo de configuracao.

### Variavel de Ambiente

Formato:

```text
env:NOME_DA_VARIAVEL
```

Com valor padrao:

```text
env:NOME_DA_VARIAVEL:valor_padrao
```

Exemplo:

```json
{
  "schema": "env:GRAPHQL_SCHEMA",
  "connectors": "env:GRAPHQL_CONNECTORS"
}
```

### AWS Systems Manager Parameter Store

Formato:

```text
ssm:/nome/do/parametro
```

Com decrypt habilitado:

```text
ssm:/nome/do/parametro:true
```

Exemplo:

```json
{
  "schema": "ssm:/graphql/dev/schema",
  "connectors": "ssm:/graphql/dev/connectors:false",
  "mock": "ssm:/graphql/dev/mock:false"
}
```

Para usar `ssm:`, o `GraphQL` precisa ser criado com `cloud.SSMContext` nos recursos.

### AWS Secrets Manager

Formatos aceitos:

```text
secret:/nome/do/segredo
secrets:/nome/do/segredo
```

Com tipo do segredo:

```text
secrets:/nome/do/segredo:json
```

Exemplo:

```json
{
  "authorization": {
    "require_token_sts": true,
    "tokenService": {
      "token_authorization_url": "env:STS_TOKEN_URL",
      "Credentials": {
        "client_id": "env:STS_CLIENT_ID",
        "client_secret": "secrets:/graphql/dev/credentials:json"
      }
    }
  }
}
```

Para usar `secret:` ou `secrets:`, o `GraphQL` precisa ser criado com `cloud.SecretsManagerContext` nos recursos.

### AWS S3

Formato:

```text
s3:REGION:BUCKET:KEY
```

Exemplo:

```json
{
  "schema": "s3:us-east-1:graphql-config:dev/schema.json",
  "connectors": "s3:us-east-1:graphql-config:dev/connectors.json",
  "mock": "s3:us-east-1:graphql-config:dev/mock.json"
}
```

O objeto S3 deve conter o JSON da configuracao esperada.

### AWS DynamoDB

Formato:

```text
dynamodb:REGION:TABLE:KEY
```

Com nomes customizados de atributos:

```text
dynamodb:REGION:TABLE:KEY:KEY_ATTRIBUTE:VALUE_ATTRIBUTE
```

Exemplo:

```json
{
  "schema": "dynamodb:us-east-1:graphql-config:schema",
  "connectors": "dynamodb:us-east-1:graphql-config:connectors:configKey:payload"
}
```

Por padrao, o item e buscado por `id = KEY` e o valor da configuracao e lido do atributo `value`. Se `VALUE_ATTRIBUTE` nao existir, o item inteiro e serializado como JSON.

## Schema

O schema e definido por JSON e convertido em tipos GraphQL.

Exemplo reduzido:

```json
{
  "types": [
    {
      "name": "Convenio",
      "fields": [
        { "name": "codigoConvenio", "type": "Int" },
        { "name": "nomeConvenio", "type": "String" }
      ]
    },
    {
      "name": "CombinedData",
      "fields": [
        { "name": "convenio", "type": "Object", "ofType": "Convenio" }
      ]
    }
  ],
  "query": {
    "name": "Query",
    "fields": [
      {
        "name": "dataSources",
        "type": "Object",
        "ofType": "CombinedData",
        "args": [
          { "name": "codigoConvenio", "type": "NonNull", "ofType": "Int" }
        ]
      }
    ]
  }
}
```

Tipos suportados:

- `Int`
- `Boolean`
- `Float`
- `String`
- `List`, usando `ofType`
- `Object`, usando `ofType`
- `NonNull`, usando `ofType`

## Conectores

Cada campo solicitado dentro da query principal e associado a um conector pelo nome do campo.

Exemplo:

```json
{
  "connectors": [
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
  ]
}
```

Campos:

- `field`: campo GraphQL resolvido pelo conector.
- `adapter`: origem dos dados. Valores atuais: `redis`, `rest`, `dynamodb`, `s3`, `rds`.
- `adapterConfig`: configuracao especifica do adapter.
- `keyPattern`: template usado para montar a chave, path ou endpoint.
- `timeoutMs`: timeout por tentativa do conector.
- `retries`: numero de novas tentativas apos falha.
- `optional`: quando `true`, uma falha nesse conector nao derruba a consulta inteira.

### Templates Dinamicos

Qualquer argumento da query pode ser usado no `keyPattern`.

Query:

```graphql
query {
  dataSources(codigoConvenio: 10341) {
    convenio {
      nomeConvenio
    }
  }
}
```

Conector:

```json
{
  "field": "convenio",
  "adapter": "redis",
  "adapterConfig": {
    "endpoint": "localhost:6379"
  },
  "keyPattern": "CVN_{codigoConvenio}"
}
```

Chave final:

```text
CVN_10341
```

## Adapters

### Redis

```json
{
  "field": "convenio",
  "adapter": "redis",
  "adapterConfig": {
    "endpoint": "localhost:6379",
    "password": ""
  },
  "keyPattern": "CVN_{codigoConvenio}"
}
```

O valor recuperado do Redis deve ser JSON e sera convertido para `map[string]interface{}`.

### REST

```json
{
  "field": "convenio",
  "adapter": "rest",
  "adapterConfig": {
    "baseUrl": "https://api.exemplo.com",
    "endpoint": "/convenios/{codigoConvenio}",
    "method": "GET",
    "headers": {
      "Accept": "application/json"
    }
  }
}
```

Para REST, se `keyPattern` nao for informado, o conector usa `adapterConfig.endpoint`.

### DynamoDB

```json
{
  "field": "convenio",
  "adapter": "dynamodb",
  "adapterConfig": {
    "region": "us-east-1",
    "table": "convenios",
    "accessKeyId": "local",
    "secretAccessKey": "local"
  },
  "keyPattern": "CVN_{codigoConvenio}"
}
```

O adapter busca o item usando a chave `id` com o valor gerado por `keyPattern`.

### S3

```json
{
  "field": "taxaFunding",
  "adapter": "s3",
  "adapterConfig": {
    "region": "us-east-1",
    "bucket": "bucket-configuracoes",
    "accessKeyId": "local",
    "secretAccessKey": "local"
  },
  "keyPattern": "funding/{codigoConvenio}.json"
}
```

O objeto S3 deve conter JSON.

### RDS

```json
{
  "field": "convenio",
  "adapter": "rds",
  "adapterConfig": {
    "driverName": "postgres",
    "dsn": "postgres://user:pass@host:5432/dbname?sslmode=require",
    "resultMode": "one"
  },
  "keyPattern": "select codigo_convenio as \"codigoConvenio\", nome_convenio as \"nomeConvenio\" from convenio where codigo_convenio = {codigoConvenio}"
}
```

Campos do `adapterConfig`:

- `driverName`: nome do driver SQL. Drivers registrados atualmente: `postgres` e `mysql`.
- `dsn`: string de conexao do banco.
- `resultMode`: `one` para retornar a primeira linha como objeto, ou `many` para retornar `{ "items": [...] }`.

Para MySQL:

```json
{
  "field": "convenios",
  "adapter": "rds",
  "adapterConfig": {
    "driverName": "mysql",
    "dsn": "user:pass@tcp(host:3306)/dbname?parseTime=true",
    "resultMode": "many"
  },
  "keyPattern": "select codigo_convenio as codigoConvenio, nome_convenio as nomeConvenio from convenio where segmento = '{segmento}'"
}
```

Observacao: o `keyPattern` do RDS e executado como SQL apos interpolacao dos argumentos. Em ambientes produtivos, prefira limitar os argumentos expostos no schema e usar fontes confiaveis de configuracao para reduzir risco de injecao.

## Mock

O mock permite testar sem acessar Redis, REST, DynamoDB ou S3.

Exemplo:

```json
{
  "status": true,
  "values": {
    "10341": {
      "convenio": {
        "codigoConvenio": 10341,
        "nomeConvenio": "Convenio de Credito Servidores Estaduais"
      },
      "limiteOperacional": {
        "percentualMaximoMargemConsignavel": 0.35
      },
      "taxaFunding": {
        "dataFunding": "2025-04-11"
      }
    }
  }
}
```

Com `status: true`, o resolver tenta retornar os dados do mock antes de consultar o conector real.

## Uso em Go

### Criando a API

```go
resources := &cloud.CloudContextList{
    cloud.SSMContext,
    cloud.SecretsManagerContext,
}

config, err := graphql.LoadConfig("examples/config/service.json")
if err != nil {
    panic(err)
}

api, err := graphql.New(config, resources, "us-east-1", "http://localhost:4566")
if err != nil {
    panic(err)
}
```

### Handler HTTP Local

```go
wrappedHandler := api.NewHandler(config.Pretty)

http.Handle(api.Config.Route, wrappedHandler)
log.Fatal(http.ListenAndServe(":8080", nil))
```

### AWS ALB

```go
wrappedHandler := api.NewHandler(config.Pretty)
adapter := graphql.WrapHandler(wrappedHandler).ToALB()

func requestHandler(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
    return adapter.ProxyWithContext(ctx, req)
}
```

### API Gateway v1

```go
adapter := graphql.WrapHandler(wrappedHandler).ToAPIGateway()
```

### API Gateway v2

```go
adapter := graphql.WrapHandler(wrappedHandler).ToAPIGatewayV2()
```

## Funcoes Publicas Principais

## Preparacao Tecnica e Feature Flags

O conector possui uma base de configuracao para ativar capacidades operacionais sem mudar o contrato principal de consulta. Essa camada concentra flags de rastreabilidade, resiliencia, MCP e mascaramento de dados sensiveis.

Exemplo:

```json
{
  "features": {
    "tracing_enabled": true,
    "mcp_enabled": false,
    "resilience_enabled": true
  },
  "security": {
    "redaction": {
      "enabled": true,
      "fields": [
        "authorization",
        "client_secret",
        "access_token",
        "refresh_token",
        "password",
        "token",
        "api_key",
        "x-api-key"
      ]
    }
  }
}
```

| Campo | Padrão | Uso |
|---|---:|---|
| `features.tracing_enabled` | `true` | Controla o middleware de trace aplicado pelo `NewHandler`. |
| `features.mcp_enabled` | `false` | Reserva para o MCP Admin Server. |
| `features.resilience_enabled` | `true` | Indica uso de retry/backoff/circuit breaker padronizados nos connectors configurados. |
| `security.redaction.enabled` | `true` | Indica que dados sensíveis devem ser mascarados em logs e diagnósticos. |

O adapter REST já mascara tokens em logs. O token continua sendo usado no header `Authorization`, mas não é impresso integralmente.

### `graphql.LoadConfig(path string) (*graphql.Config, error)`

Carrega um arquivo JSON de configuracao e resolve paths locais relativos ao diretorio do arquivo.

```go
config, err := graphql.LoadConfig("examples/config/service.json")
```

### `graphql.New(config, resources, region, endpoint) (*graphql.GraphQL, error)`

Cria o contexto cloud, carrega conectores, cria resolver, monta schema e configura autorizacao quando habilitada.

```go
api, err := graphql.New(config, resources, "us-east-1", "http://localhost:4566")
```

### `(*GraphQL).NewHandler(pretty bool, middlewares ...graphql.Middleware) http.Handler`

Cria o handler GraphQL. Os middlewares de rastreabilidade e tempo de execucao sao aplicados automaticamente.

```go
handler := api.NewHandler(true)
```

O handler entende o header W3C `traceparent`. Quando a requisicao ja chega com esse header, o conector preserva o `trace_id` e usa esse contexto nas chamadas feitas pelos connectors. Quando a requisicao nao possui trace, o conector cria um novo identificador.

Exemplo de requisicao:

```http
POST /graphql HTTP/1.1
Content-Type: application/json
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

Os adapters REST propagam os headers abaixo para as APIs integradas:

```http
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-9f3a4c7b12e64010-01
X-Trace-ID: 4bf92f3577b34da6a3ce929d0e0e4736
X-Span-ID: 9f3a4c7b12e64010
X-Parent-Span-ID: 00f067aa0ba902b7
```

Esse comportamento permite acompanhar uma consulta GraphQL do workflow ate a API externa consultada pelo connector.

### `graphql.WrapHandler(h http.Handler)`

Empacota um `http.Handler` para conversao em adapters AWS.

```go
adapter := graphql.WrapHandler(handler).ToALB()
```

### `graphql.FromString(value string) *graphql.Path`

Converte um path inline em uma estrutura de captura.

```go
path := graphql.FromString("ssm:/graphql/dev/schema:false")
value, err := path.GetValue(cloudContext)
```

### `graphql.IsPath(value string) bool`

Indica se uma string usa algum prefixo de captura suportado.

```go
ok := graphql.IsPath("local:schema.json")
```

## Headers de Resposta

Toda resposta GraphQL criada com `api.NewHandler(...)` recebe:

```http
x-graphql-elapsed-time: 4
X-Trace-ID: 4bf92f3577b34da6a3ce929d0e0e4736
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

O header `x-graphql-elapsed-time` informa o tempo total de processamento da requisicao em milissegundos. Os headers `X-Trace-ID` e `traceparent` permitem correlacionar a chamada com o workflow, as APIs integradas e o `custom-business-metrics`.

## Rastreabilidade Distribuida

O conector suporta rastreabilidade distribuida para que uma chamada originada no `routing-slip-pattern` continue identificavel dentro do GraphQL connector e nas APIs consultadas por ele.

Fluxo:

```mermaid
sequenceDiagram
    participant Workflow as routing-slip-pattern
    participant GraphQL as go-graphql-connector
    participant REST as API REST integrada

    Workflow->>GraphQL: query GraphQL + traceparent
    GraphQL->>GraphQL: preserva trace_id e cria contexto
    GraphQL->>REST: connector REST + traceparent filho
    REST-->>GraphQL: resposta da API
    GraphQL-->>Workflow: resposta + X-Trace-ID
```

Pontos importantes:

- o `trace_id` permanece o mesmo durante toda a jornada;
- cada connector REST recebe um novo `span_id`;
- o handler GraphQL e cada connector criam spans usando a API do OpenTelemetry;
- o `traceparent` e propagado automaticamente pelo adapter REST;
- CORS permite os headers `traceparent`, `X-Trace-ID`, `X-Span-ID` e `X-Parent-Span-ID`;
- tokens STS continuam funcionando normalmente e nao sao expostos nos headers de rastreio.

Exemplo de uso com `curl`:

```bash
curl --request POST \
  --url http://localhost:8090/graphql \
  --header 'content-type: application/json' \
  --header 'traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01' \
  --data '{"query":"query { dataSources(id: \"PED-1001\") { order { id status } } }"}'
```

## Resiliencia por Connector

A resiliencia configuravel por connector permite lidar com falhas transitorias de APIs, timeouts e instabilidade sem jogar essa responsabilidade para o workflow.

Exemplo:

```json
{
  "field": "order",
  "adapter": "rest",
  "adapterConfig": {
    "baseUrl": "https://mock.raysouz.studio",
    "endpoint": "/orders/{id}",
    "method": "GET"
  },
  "keyPattern": "/orders/{id}",
  "timeoutMs": 1500,
  "retries": 2,
  "resilience": {
    "backoff": "exponential",
    "initial_interval_ms": 100,
    "max_interval_ms": 1000,
    "jitter": true,
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 5,
      "open_timeout_ms": 30000
    }
  }
}
```

Campos principais:

| Campo | Uso |
|---|---|
| `timeoutMs` | Timeout por tentativa do connector. |
| `retries` | Quantidade de novas tentativas após a primeira chamada. |
| `resilience.backoff` | `exponential`, `fixed` ou `none`. |
| `resilience.initial_interval_ms` | Espera inicial entre tentativas. |
| `resilience.max_interval_ms` | Limite máximo do backoff. |
| `resilience.jitter` | Adiciona variação ao backoff para evitar rajadas. |
| `resilience.circuit_breaker.enabled` | Liga o circuit breaker do connector. |
| `resilience.circuit_breaker.failure_threshold` | Falhas consecutivas necessárias para abrir o circuito. |
| `resilience.circuit_breaker.open_timeout_ms` | Tempo em que o circuito permanece aberto. |

Classificação de erros:

| Classe | Quando ocorre | Retry? |
|---|---|---|
| `retryable` | Erros transitórios genéricos. | Sim |
| `timeout` | Deadline ou timeout da chamada. | Sim |
| `auth_error` | Erros de autenticação/autorização. | Não |
| `non_retryable` | Configuração inválida, argumento ausente ou erro funcional da resposta. | Não |
| `circuit_open` | Circuit breaker aberto. | Não |

O estado do circuit breaker fica disponível programaticamente:

```go
states := api.ConnectorCircuitStates()
```

Cada chamada de connector também registra atributos de span como `connector.field`, `connector.attempt`, `connector.retries`, `connector.error_class` e `connector.circuit_breaker.enabled`.

## Validacao

Execute:

```bash
go test ./...
```

Para checar concorrencia nos pacotes principais:

```bash
go test -race ./internal/graph ./internal/graph/connectors ./graphql
```

## Integracao com State Store do Routing Slip

O state store persistente do `routing-slip-pattern` guarda snapshots de execucao e permite retomar workflows a partir do ponto salvo. O `go-graphql-connector` nao precisa persistir snapshots do workflow, mas participa desse desenho como fonte de enriquecimento reexecutavel e rastreavel.

Quando um step `graphql_enrich` consulta o conector, o resultado enriquecido e salvo dentro do snapshot do workflow. Se a execucao falhar depois da consulta, o reprocessamento pode continuar a partir do cursor salvo sem repetir etapas anteriores ja concluidas. Caso o cursor volte manualmente para uma etapa ja marcada como `success`, a idempotencia do runtime pode registrar `idempotent_skip` e seguir o fluxo.

Boas praticas para usar o conector com state store:

- propagar `traceparent`, `X-Trace-ID` e `X-Correlation-ID` nas chamadas;
- manter `timeoutMs`, `retries` e circuit breaker configurados em connectors que consultam APIs externas;
- usar respostas deterministicas quando o enriquecimento sera reutilizado apos reprocessamento;
- evitar side effects em connectors de leitura, deixando efeitos externos para handlers explicitos do workflow.

Exemplo de step:

```yaml
- id: carregar-produto
  name: graphql_enrich
  params:
    endpoint: http://go-graphql-connector:8090/graphql
    target: catalogo
    result_path: dataSources
    required: true
```

O snapshot persistido pelo workflow guarda o payload apos esse enriquecimento, junto com cursor, historico, `trace_id` e estado granular da etapa.

## MCP Admin

O MCP Admin define uma camada para expor capacidades administrativas do `go-graphql-connector` sem vazar segredos ou acoplar o Studio diretamente aos arquivos de configuracao.

Tools previstas para o MCP Admin:

| Tool | Objetivo |
|---|---|
| `list_connectors` | Lista connectors configurados. |
| `describe_connector` | Mostra adapter, timeout, resiliencia e schema esperado. |
| `test_connector` | Testa um connector isoladamente com entrada controlada. |
| `execute_graphql_query` | Executa uma query diagnostica controlada. |
| `get_token_status` | Mostra status do token gerenciado sem retornar o token. |
| `get_circuit_state` | Mostra estado dos circuit breakers por connector. |
| `validate_config` | Valida `schema.json`, `connectors.json` e `service.json`. |

Regras de seguranca:

- nunca retornar `client_secret`, access token, authorization header ou certificados;
- mascarar campos sensiveis antes de responder tools;
- permitir allowlist de connectors testaveis;
- manter tools de execucao desativaveis em producao;
- preservar propagacao de `traceparent`, `X-Trace-ID` e `X-Correlation-ID` nas chamadas diagnosticas.

O contrato deve alinhar Studio, agentes e automacoes de suporte. A execucao real das tools administrativas deve respeitar as mesmas politicas de resiliencia, token gerenciado e redaction ja existentes no conector.

## Planner MCP e Integracoes GraphQL

O planner assistido por MCP do `routing-slip-pattern` pode sugerir steps `graphql_enrich` ou `rest_call` quando identifica enriquecimento, consulta a API ou necessidade de compor dados externos.

Para o `go-graphql-connector`, isso significa que a configuracao de connectors deve continuar explicita e auditavel. O planner pode gerar um rascunho de step, mas a query, variaveis, `target` e `result_path` devem ser revisados antes da execucao.

Exemplo de saida esperada pelo planner:

```yaml
- id: carregar-contexto
  name: graphql_enrich
  params:
    endpoint: http://go-graphql-connector:8090/graphql
    query: "query { dataSources { status } }"
    result_path: dataSources
    target: contexto
    required: true
```

Recomendacoes:

- validar a query com `validate_config` ou consulta manual antes de usar em producao;
- manter `timeoutMs`, `retries` e circuit breaker nos connectors envolvidos;
- revisar tokens, certificados e redaction;
- preferir consultas sem side effect para etapas de enriquecimento;
- documentar quais campos enriquecidos entram no payload persistido do workflow.
