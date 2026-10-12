# 002 — Resolve Performance Test: Horizontal Scaling with Nginx

## 1. Introduction

In the previous performance test, documented in [`001-resolve-performance-test.md`](./001-resolve-performance-test.md), we evaluated the `url-service` resolve endpoint with Redis caching enabled.

The final results were:

| RPS | P50 | P90 | P99 | Error Rate |
|---:|---:|---:|---:|---:|
| 1,000 | 762.27 µs | 986.70 µs | 1.48 ms | 0.00% |
| 2,000 | 693.66 µs | 914.98 µs | 1.57 ms | 0.00% |
| 3,000 | 681.99 µs | 926.78 µs | 3.31 ms | 0.00% |
| 4,000 | 667.85 µs | 1.14 ms | 53.40 ms | 0.00% |
| 5,000 | 1.07 ms | 15.32 ms | 37.93 ms | 0.00% |
| 6,000 | 38.34 ms | 134.03 ms | 204.83 ms | 0.00% |
| 7,000 | 807.38 ms | 1.48 s | 1.76 s | 0.00% |

Although Redis caching improved resolve performance, increasing the request rate eventually caused a significant rise in latency. The `url-service` instances approached CPU saturation, limiting the system's ability to handle additional traffic efficiently.

Rather than relying on a single service instance, the next step is to introduce **horizontal scaling**: run multiple instances of `url-service` and distribute incoming requests between them through a load balancer.

This article investigates whether running three service instances behind Nginx can improve the resolve endpoint's performance under higher load.

## 2. Architecture

The updated request flow is:

```text
                 k6 Load Test
                       |
                       v
                Nginx Gateway
                (Load Balancer)
                       |
          +------------+------------+
          |            |            |
          v            v            v
     url-service-1 url-service-2 url-service-3
          |            |            |
          +------------+------------+
                       |
                       v
                  PostgreSQL
                  Redis Cache
```

Nginx acts as the entry point and distributes incoming requests across three `url-service` instances. The service instances continue to use the existing PostgreSQL and Redis infrastructure.

## 3. Nginx Gateway Implementation

An Nginx gateway was introduced to distribute incoming requests across three `url-service` instances.

The configuration includes the following settings:

- **Worker Configuration:** Set `worker_processes auto` and configured worker connection limits.
- **Connection Capacity:** Set `worker_connections 20000` and `worker_rlimit_nofile 30000`.
- **Load Balancing:** Configured an upstream group containing three backend instances.
- **Keepalive Connections:** Enabled HTTP/1.1 and upstream keepalive to reuse backend connections.
- **Logging:** Disabled access logs to reduce overhead during performance testing.
- **Failure Handling:** Configured `max_fails` and `fail_timeout` for upstream servers.
- **File Transfer:** Enabled `sendfile`.

These settings establish the Nginx gateway configuration used for the horizontal-scaling experiment.

## 4. Performance Testing Objective

The next step is to repeat the resolve performance test with the Nginx gateway and three `url-service` instances.

The goal is to compare the new results against the baseline from `001-resolve-performance-test.md` and determine whether horizontal scaling improves latency and allows the system to handle higher request rates.

No performance improvement is assumed in advance; the results of the new tests will determine whether the additional instances provide a measurable benefit.

## 5. Performance Test Results

The resolve endpoint was tested with the Nginx gateway distributing traffic across three `url-service` instances. The test covered request rates from 1,000 to 10,000 RPS.

| RPS | P50 | P90 | P99 | Error Rate |
|---:|---:|---:|---:|---:|
| 1,000 | 977.73 µs | 1.23 ms | 1.63 ms | 0.00% |
| 2,000 | 896.07 µs | 1.12 ms | 1.68 ms | 0.00% |
| 3,000 | 825.26 µs | 966.77 µs | 2.37 ms | 0.00% |
| 4,000 | 829.36 µs | 1.03 ms | 6.80 ms | 0.00% |
| 5,000 | 836.36 µs | 1.03 ms | 10.99 ms | 0.00% |
| 6,000 | 846.96 µs | 1.08 ms | 27.99 ms | 0.00% |
| 7,000 | 865.57 µs | 1.22 ms | 46.88 ms | 0.00% |
| 8,000 | 943.17 µs | 9.06 ms | 66.59 ms | 0.00% |
| 9,000 | 33.73 ms | 95.76 ms | 156.76 ms | 0.00% |
| 10,000 | 77.22 ms | 136.20 ms | 170.34 ms | 0.00% |

## 6. Analysis

The results show that the system successfully handled a target request rate of **10,000 RPS**, with an error rate of 0.00% throughout the tested range.

Compared with the previous test, which used a single `url-service` instance, the three-instance setup allowed us to test at higher request rates while maintaining relatively low latency across much of the range.

At 7,000 RPS, the P99 latency decreased from **1.76 seconds** in test `001` to **46.88 ms** in this test. At 10,000 RPS, P99 latency was **170.34 ms**, and the test still reported no errors.

However, latency began increasing more sharply at 9,000 RPS, particularly in P50 and P90. This suggests that the system was approaching another capacity limit, even though the measured error rate remained at zero.

These results indicate that adding service instances behind Nginx significantly increased the tested throughput range. They also provide a basis for investigating whether the system can handle request rates beyond 10,000 RPS.

## 7. Conclusion

Horizontal scaling with three `url-service` instances behind Nginx enabled the resolve endpoint to reach **10,000 RPS with a reported error rate of 0.00%**.

The next step is to continue testing at higher request rates to identify the system's new performance limits and observe how latency and error rates change under additional load.