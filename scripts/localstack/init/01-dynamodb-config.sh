#!/usr/bin/env bash
set -euo pipefail

TABLE_NAME="graphql-config"
CONFIG_DIR="/opt/go-graphql-connector/config"
CACHE_CLUSTER_ID="graphql-redis"
USE_LOCALSTACK_ELASTICACHE="${USE_LOCALSTACK_ELASTICACHE:-false}"

awslocal dynamodb create-table \
  --table-name "${TABLE_NAME}" \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST >/dev/null 2>&1 || true

awslocal dynamodb wait table-exists --table-name "${TABLE_NAME}"

if [[ "${USE_LOCALSTACK_ELASTICACHE}" == "true" ]]; then
  awslocal elasticache create-cache-cluster \
    --cache-cluster-id "${CACHE_CLUSTER_ID}" \
    --cache-node-type cache.t2.micro \
    --engine redis \
    --num-cache-nodes 1 >/dev/null
fi

python3 - <<'PY'
import json
import os
import pathlib
import socket
import subprocess
import time

table_name = "graphql-config"
config_dir = pathlib.Path("/opt/go-graphql-connector/config")
cache_cluster_id = "graphql-redis"
use_localstack_elasticache = os.environ.get("USE_LOCALSTACK_ELASTICACHE", "false").lower() == "true"

def put_config(key, value):
    item = {
        "id": {"S": key},
        "value": {"S": value},
    }
    subprocess.run(
        [
            "awslocal",
            "dynamodb",
            "put-item",
            "--table-name",
            table_name,
            "--item",
            json.dumps(item),
        ],
        check=True,
    )

def get_redis_endpoint():
    for _ in range(60):
        output = subprocess.check_output(
            [
                "awslocal",
                "elasticache",
                "describe-cache-clusters",
                "--cache-cluster-id",
                cache_cluster_id,
                "--show-cache-node-info",
            ],
            text=True,
        )
        payload = json.loads(output)
        clusters = payload.get("CacheClusters", [])
        if clusters:
            nodes = clusters[0].get("CacheNodes", [])
            if nodes and nodes[0].get("Endpoint"):
                endpoint = nodes[0]["Endpoint"]
                return endpoint["Address"], int(endpoint["Port"])
        time.sleep(1)
    raise RuntimeError("ElastiCache Redis endpoint was not available")

def redis_command(host, port, *parts):
    encoded = [str(part).encode("utf-8") for part in parts]
    payload = b"*" + str(len(encoded)).encode("ascii") + b"\r\n"
    for part in encoded:
        payload += b"$" + str(len(part)).encode("ascii") + b"\r\n" + part + b"\r\n"

    with socket.create_connection((host, port), timeout=5) as sock:
        sock.sendall(payload)
        response = sock.recv(4096)
        if response.startswith(b"-"):
            raise RuntimeError(response.decode("utf-8", errors="replace"))

if use_localstack_elasticache:
    redis_host, redis_port = get_redis_endpoint()
    redis_endpoint = f"{redis_host}:{redis_port}"
    redis_origin = "elasticache"
else:
    redis_host, redis_port = "go-graphql-connector-elasticache-redis", 6379
    redis_endpoint = "localhost:6379"
    redis_origin = "elasticache-compatible-redis"

redis_command(
    redis_host,
    redis_port,
    "SET",
    "CVN_10341",
    json.dumps(
        {
            "codigoConvenio": 10341,
            "nomeConvenio": "Convênio de Crédito Servidores Estaduais",
            "origem": redis_origin,
        },
        ensure_ascii=False,
    ),
)

connectors = {
    "connectors": [
        {
            "field": "convenio",
            "adapter": "redis",
            "adapterConfig": {
                "endpoint": redis_endpoint,
                "password": "",
            },
            "keyPattern": "CVN_{codigoConvenio}",
            "timeoutMs": 500,
            "retries": 1,
        },
        {
            "field": "limiteOperacional",
            "adapter": "rest",
            "adapterConfig": {
                "baseUrl": "http://localhost:8090",
                "endpoint": "/limites/{codigoConvenio}",
                "method": "GET",
                "headers": {
                    "Accept": "application/json",
                },
            },
            "timeoutMs": 1000,
            "retries": 1,
        },
    ]
}

put_config("schema", (config_dir / "schema.json").read_text(encoding="utf-8"))
put_config("connectors", json.dumps(connectors, ensure_ascii=False))
print(f"Loaded GraphQL config into DynamoDB table {table_name} using ElastiCache endpoint {redis_endpoint}")
PY
