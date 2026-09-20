# Test Contract

## Version Under Test

- All tests are executed against version **v1.4.1**. All recorded results below correspond to this exact version.

## Test Environment (Containerized)

To standardize the tests, all service resources are containerized in Docker:

| Service           | cpus | memory | Host Ports                  | Volume          |
|-------------------|------|--------|-----------------------------|-----------------|
| postgres          | 1.0  | 1G     | 15432:5432                  | postgres_perf   |
| redis             | 0.5  | 512M   | 16379:6379                  | redis_perf      |
| rabbitmq          | 0.5  | 512M   | 25672:5672 + 15673:15672    | rabbitmq_perf   |
| url-service       | 1.5  | 1G     | 8081:8080                   | —               |
| analytics-service | 0.5  | 512M   | (no port)                   | —               |
| k6                | no limit | —   | —                           | —               |

## Further Reading

See branch **`perf/test-environment`** for the load tests, related test articles, and conclusions.