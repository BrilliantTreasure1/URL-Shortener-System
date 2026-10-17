# Performance Test Environment

An **isolated, containerized** environment for load-testing the URL shortener. It runs as its
own Docker Compose project (`url-shortener-perf`) with its own volumes, so experiments never
touch development data.

The current setup runs **three `url-service` replicas behind an Nginx load balancer**, with
[k6](https://k6.io/) as the load generator and [Jaeger](https://www.jaegertracing.io/) for
request tracing.

## Architecture

```text
                 k6 Load Test
                       │
                       ▼
                Nginx Gateway            (:8081)
               (Load Balancer)
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
  url-service-1  url-service-2  url-service-3
        │              │              │
        └──────────────┼──────────────┘
                       │
                       ▼
              PostgreSQL  ·  Redis
                       │
                       └── OTLP traces → Jaeger (:16686)
```

## Services

Every service is pinned with compose-native `cpus` / `mem_limit` caps, so the resource budget
of each experiment is explicit and reproducible. The full contract is in
[`test_contract.md`](test_contract.md).

| Service (container)                       | cpus | memory | Host port(s)            |
|-------------------------------------------|------|--------|-------------------------|
| postgres (`url-shortener-perf-db`)        | 1.0  | 1g     | 15432:5432              |
| redis (`…-redis`)                         | 0.5  | 512m   | 16379:6379              |
| rabbitmq (`…-rabbitmq`)                   | 0.5  | 512m   | 25672:5672, 15673:15672 |
| jaeger (`…-jaeger`)                       | 0.5  | 512m   | 16686, 4317, 4318       |
| url-service ×3 (`…-url-service-{1,2,3}`)  | 1.5  | 1g     | — (behind gateway)      |
| analytics-service (`…-analytics-service`) | 0.5  | 512m   | —                       |
| gateway (`…-gateway`, nginx:1.27-alpine)  | 2.0  | 512m   | 8081:80                 |
| k6 (`…-k6`)                               | —    | —      | —                       |

## The load test

[`resolve/resolve_load_test.js`](resolve/resolve_load_test.js) drives a `constant-arrival-rate`
scenario against the resolve endpoint **through the gateway**:

```js
const res = http.get('http://gateway/1J9Mt4');   // expect 302
```

Redirects are intentionally not followed (`maxRedirects: 0`), so the measured response is the
service's own `302` rather than the target site's page.

## Running

```bash
cd performance
docker compose up -d --build --scale url-service=3
docker compose run --rm k6
```

- Jaeger UI: <http://localhost:16686>
- Through the load balancer: <http://localhost:8081>

To probe a different rate, edit `rate` (and `preAllocatedVUs` / `maxVUs`) in the k6 script and
re-run the `k6` service.

## Latest baseline (v1.4.1)

Three replicas behind Nginx sustained **10,000 RPS at a 0.00% error rate**.

| RPS | P50 | P90 | P99 | Errors |
|---:|---:|---:|---:|---:|
| 7,000 | 865.57 µs | 1.22 ms | 46.88 ms | 0.00% |
| 10,000 | 77.22 ms | 136.20 ms | 170.34 ms | 0.00% |

At 7,000 RPS the P99 dropped from **1.76 s** with a single instance (`001`) to **46.88 ms** here
(`002`).

## Documentation

- [`test_contract.md`](test_contract.md) — version under test and per-service resource limits
- [`resolve/doc/001-resolve-performance-test.md`](resolve/doc/001-resolve-performance-test.md) —
  single-instance baseline (1,000–7,000 RPS)
- [`resolve/doc/002-resolve-performance-test.md`](resolve/doc/002-resolve-performance-test.md) —
  horizontal scaling with Nginx (1,000–10,000 RPS)

## Notes

- The gateway's upstream targets the three replicas **by compose service name**
  (`url-shortener-perf-url-service-1…3`), so `--scale url-service=3` must be used.
- This stack is for measurement only; it is independent of the production-oriented
  [`deployment/`](../deployment) compose file.
