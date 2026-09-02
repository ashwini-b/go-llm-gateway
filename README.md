# llm-gateway

A lightweight, OpenAI-compatible LLM gateway written in Go. It sits in front of one or more LLM providers and handles auth, routing, rate limiting, caching, retries/failover, and observability — similar in spirit to LiteLLM or Portkey, scoped down to a focused v1.

## Features (v1)

- **OpenAI-compatible API** — drop-in `/v1/chat/completions` endpoint
- **Provider abstraction** — `Provider` interface with an Ollama implementation out of the box
- **Config-driven routing** — map model names to one or more upstream providers via YAML
- **API-key auth** — Bearer token middleware, keys defined in config
- **Structured logging** — JSON logs via `log/slog`
- **Per-key rate limiting** — token-bucket limiting (`golang.org/x/time/rate`) per API key
- **In-memory response cache** — TTL-based cache to skip repeat calls to slow upstreams
- **Retries + failover** — automatic failover across multiple providers registered for the same model
- **Metrics** — Prometheus-compatible `/metrics` endpoint alongside `/healthz`
- **Dockerized** — multi-stage Dockerfile and a `docker-compose.yml` that runs the gateway against two independent Ollama containers to exercise failover locally

## Architecture

```
                     ┌─────────────┐
 client ── HTTP ──▶  │  llm-gateway │
                     └──────┬──────┘
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                    ▼
   auth middleware   rate limiter (per key)   response cache
        │
        ▼
   provider registry ── model name ──▶ [provider, provider, ...]
        │
        ▼
   Provider.Complete()  ── retries/failover across registered providers
        │
        ▼
   upstream (e.g. Ollama)
```

Requests to `/v1/chat/completions` are authenticated, rate-limited per key, checked against the cache, then routed by model name through the provider registry. If a model has multiple providers configured, the gateway retries against the next one on failure.

## Getting Started

### Prerequisites

- Go 1.22+
- [Ollama](https://ollama.com) running locally (or via Docker Compose, see below)
- Docker + Docker Compose (optional, for the containerized setup)

### Run locally

```bash
# pull a model for Ollama to serve
ollama pull llama3

# run the gateway
go run ./cmd
```

The server starts on the port configured in `config.yaml` (default shown below).

### Run with Docker Compose

```bash
docker compose up --build
```

This starts the gateway plus two separate Ollama containers (mirroring a real failover setup). Model pulls are **manual**, not automatic on startup — pull the model into each Ollama container before sending requests:

```bash
docker exec -it <ollama-container-name> ollama pull llama3
```

## Configuration

Configuration lives in `config.yaml` (or `config.docker.yaml` for the Compose setup):

```yaml
server:
  port: 8080

auth:
  api_keys:
    - "your-api-key-here"

models:
  - name: llama3
    providers:
      - provider: ollama
        base_url: http://localhost:11434
      - provider: ollama
        base_url: http://localhost:11435   # second instance, used on failover

log:
  level: info

rate_limit:
  rps: 5
  burst: 10

cache:
  enabled: true
  ttl_seconds: 300
```

- **`models`** maps a model name to one or more providers. Listing more than one provider for a model enables retry/failover between them.
- **`auth.api_keys`** are the Bearer tokens accepted by the auth middleware.
- **`rate_limit`** is applied per API key.
- **`cache`** controls the in-memory TTL response cache.

## API

### `POST /v1/chat/completions`

OpenAI-compatible request/response shape.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### `GET /healthz`

Liveness check.

### `GET /metrics`

Prometheus metrics endpoint.

## Testing

```bash
go test ./...
```

## Roadmap (post-v1)

Planned but out of scope for v1:

- Emit usage/cost events to Kafka
- Swap the in-memory cache for Redis
- Postgres-backed usage/cost history with a small dashboard
- Multi-provider load balancing (beyond failover)
- Deploy to AWS (ECS/EKS) or Kubernetes with a Helm chart

## License

MIT
