## Test Objective

The objective of this experiment is to evaluate the performance of the **resolve path without cache** and understand how the system performs when Redis is unavailable and no cached data can be used.

During the initial test, I observed that for every incoming request, the application attempted to establish a connection to Redis. The connection logic retried the operation **five times** before failing. This behavior introduced significant delays and caused requests to accumulate in the processing queue.

As a result, a considerable number of requests either **failed** or took approximately **one minute** to complete.

To address this issue, I made several changes to the Redis connection behavior, including implementing **fail-fast behavior** during connection attempts and introducing a **circuit breaker**.

These changes significantly improved the request processing behavior. The processing speed increased, the failure rate at the beginning of the test decreased noticeably, and the **maximum request processing time was reduced to less than three seconds**.


## RPS Capacity Test

In the next stage, I started using **k6** to determine the system's RPS capacity. I gradually increased the request rate from **100 RPS to 700 RPS** and recorded the following results:

| RPS |     P50 |     P90 |     P99 | Error Rate |
| --: | ------: | ------: | ------: | ---------: |
| 100 | 1.64 ms | 1.96 ms | 2.33 ms |      0.00% |
| 200 | 1.08 ms | 1.44 ms | 1.83 ms |      0.00% |
| 300 |  5.74 s | 59.99 s | 59.99 s |     22.89% |
| 400 |  6.11 s | 59.99 s | 59.99 s |     28.60% |
| 500 |  9.43 s | 59.99 s | 59.99 s |     46.24% |
| 700 | 59.99 s | 59.99 s | 59.99 s |     63.82% |

As shown in the results, the system performed normally at **100 and 200 RPS**, with no errors and very low latency.

However, starting at **300 RPS**, the error rate increased significantly, while **P90 and P99 latency also increased sharply**. This indicated that the system was reaching a bottleneck under higher load.

### Bottleneck Investigation

To identify the source of the bottleneck, I investigated the request execution path using **Jaeger** and Docker resource metrics.

In Jaeger, I observed that for a request with a total processing time of approximately **30 seconds**, at least **29 seconds were spent inside PostgreSQL**.

I then checked the Docker resource usage and observed that the PostgreSQL container's CPU utilization had reached **100%** during the test.

Based on these observations, PostgreSQL became the primary suspect for the bottleneck.

I then investigated the PostgreSQL layer in more detail. The database indexes were correctly configured, and the query itself was performing as expected. However, I found that the **database connection pool had not been properly configured**, allowing an unbounded number of database connections to be opened.

Based on this finding, I decided to properly configure the **database connection pool** and repeat the performance tests to determine its impact on system performance.

## Database Connection Pool Optimization

Based on the bottleneck investigation, I configured the PostgreSQL database connection pool with explicit limits:

```go
db.SetMaxOpenConns(20)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(2 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
```

After applying these changes, I repeated the same RPS capacity test under the same testing conditions.

The results changed significantly:

| RPS |       P50 |     P90 |      P99 | Error Rate |
| --: | --------: | ------: | -------: | ---------: |
| 100 |   1.59 ms | 1.99 ms |      2 s |      0.00% |
| 200 |   1.13 ms | 1.51 ms |  1.93 ms |      0.00% |
| 300 | 961.67 µs | 1.23 ms |  1.60 ms |      0.00% |
| 400 | 925.27 µs | 1.19 ms |  1.80 ms |      0.00% |
| 500 | 868.66 µs | 1.12 ms |  1.50 ms |      0.00% |
| 600 | 849.96 µs | 1.09 ms |  2.20 ms |      0.00% |
| 700 | 826.06 µs | 1.08 ms | 13.52 ms |      0.00% |

The results show a significant improvement compared to the previous test. The system maintained a **0.00% error rate across all tested load levels**, including 700 RPS.

Latency also remained in the millisecond range across the tested RPS levels. Even at **700 RPS**, P50 and P90 remained below 2 ms, while P99 increased to 13.52 ms.

This indicates that properly configuring the database connection pool had a significant impact on the system's behavior under load and eliminated the severe latency and failure pattern observed in the previous test.


## Next RPS Capacity Test

After reviewing the CPU utilization at **700 RPS**, I observed that the CPU usage was only around **20%**. This indicates that the system had not yet reached significant CPU saturation at this load level.

Since the system was still processing requests with a **0.00% error rate** and low latency at 700 RPS, I decided to continue increasing the request rate to identify the new performance limit before introducing the cache layer.

For the next round of testing, I will start at **1,000 RPS** and gradually increase the load from there while monitoring:

* P50, P90, and P99 latency
* Error rate
* CPU utilization
* PostgreSQL CPU utilization
* Database connection pool behavior
* Request throughput

The goal of this stage is to determine the system's capacity **without cache** and identify the next bottleneck before introducing Redis into the resolve path.


## Connection Pool Configuration Validation

Before starting the next capacity test at **1,000 RPS**, I decided to validate the database connection pool configuration in more detail.

The connection pool was configured with a maximum of **20 open connections** and **10 idle connections**. Rather than treating these values as arbitrary configuration parameters, I monitored the pool statistics during load testing to verify how the application was actually using the available connections.

The pool statistics were logged during the test, resulting in observations such as:

```text
Open=10  InUse=1  Idle=9  MaxOpen=20
```

The statistics showed that, at the observed point in time, the application had **10 open connections**, but only **1 connection was actively in use**, while **9 connections were idle**.

The statistics were collected periodically rather than continuously, so they represent a snapshot of the pool at the time of logging. However, the observations indicate that the configured pool capacity was not being approached under the tested workload.

This validation provided evidence that the connection pool configuration was not selected randomly. The configured limits were based on the observed behavior of the application and provided sufficient connection capacity for the current workload.

With the connection pool configuration validated, the next step is to continue the capacity test starting from **1,000 RPS** and determine where the next performance bottleneck appears.


## Extended Capacity Test

After validating the database connection pool configuration, I continued the capacity testing from **1,000 RPS to 8,000 RPS** without introducing the cache layer.

The following results were recorded:

|   RPS |       P50 |       P90 |       P99 | Error Rate |
| ----: | --------: | --------: | --------: | ---------: |
| 1,000 | 770.37 µs | 995.39 µs |  12.89 ms |      0.00% |
| 2,000 | 741.96 µs |   1.30 ms |  88.88 ms |      0.00% |
| 3,000 |   2.10 ms |  77.30 ms | 170.77 ms |      0.00% |
| 4,000 |  28.60 ms | 107.03 ms | 212.59 ms |      0.00% |
| 5,000 |  49.57 ms | 178.25 ms | 430.57 ms |      0.00% |
| 6,000 | 564.21 ms |    2.45 s |    5.69 s |      0.00% |
| 7,000 | 905.60 ms |    3.52 s |    7.48 s |      0.00% |
| 8,000 |    1.01 s |    3.74 s |    7.78 s |      0.00% |

### Observations

The system maintained a **0.00% error rate throughout the entire test range**, including at 8,000 RPS.

However, latency increased progressively as the request rate increased. The change became particularly significant from **3,000 RPS onward**, with P90 increasing from 77.30 ms at 3,000 RPS to 3.74 seconds at 8,000 RPS.

At the same time, I observed that the database connection pool reached its configured maximum of **20 open connections**. This indicates that, under the higher workloads, the application was able to utilize the full configured database connection capacity.

In addition to connection pool saturation, the **URL service's CPU and memory utilization also reached saturation** during the higher-load tests.

These observations indicate that the system had moved beyond the low-load regime observed in the earlier tests and was now approaching resource limits at the application layer. Further investigation is required to determine the relative contribution of the database connection pool, application CPU, memory, and other components before introducing the cache layer.

At this stage, the test provides a baseline for the **uncached resolve path under high load**, which can later be compared against the same workload after introducing Redis caching.


## Cache Impact Evaluation

After reaching the resource saturation point of the `url-service` during the uncached tests, I was no longer able to meaningfully increase the request rate beyond **8,000 RPS** in the current test environment.

I therefore introduced Redis caching and repeated the performance tests from **1,000 RPS to 7,000 RPS** to measure the impact of caching on the resolve path.

The results were:

|   RPS |       P50 |       P90 |       P99 | Error Rate |
| ----: | --------: | --------: | --------: | ---------: |
| 1,000 | 762.27 µs | 986.70 µs |   1.48 ms |      0.00% |
| 2,000 | 693.66 µs | 914.98 µs |   1.57 ms |      0.00% |
| 3,000 | 681.99 µs | 926.78 µs |   3.31 ms |      0.00% |
| 4,000 | 667.85 µs |   1.14 ms |  53.40 ms |      0.00% |
| 5,000 |   1.07 ms |  15.32 ms |  37.93 ms |      0.00% |
| 6,000 |  38.34 ms | 134.03 ms | 204.83 ms |      0.00% |
| 7,000 | 807.38 ms |    1.48 s |    1.76 s |      0.00% |

Compared with the previous uncached tests, Redis caching significantly improved the latency characteristics of the resolve path, particularly at higher request rates.

For example, at **6,000 RPS**, the uncached configuration reached:

* P50: **564.21 ms**
* P90: **2.45 s**
* P99: **5.69 s**

With caching enabled at the same load:

* P50: **38.34 ms**
* P90: **134.03 ms**
* P99: **204.83 ms**

At **7,000 RPS**, the difference was also significant. The uncached configuration reached a P99 of **7.48 s**, while the cached configuration reduced P99 to **1.76 s**.

The error rate remained at **0.00%** throughout both test ranges.

## Conclusion and Next Steps

The performance investigation was conducted in several stages.

First, the **PostgreSQL layer was measured and investigated** after the initial performance tests showed significant latency and failures under higher load. The investigation showed that the database layer was a major part of the request processing path, and the database connection pool was not properly configured.

The connection pool was then optimized by introducing explicit limits for open and idle connections and configuring connection lifetime and idle time. After this change, the system's performance improved significantly, allowing the test workload to increase from the previous **700 RPS range to 8,000 RPS** without errors.

Next, **Redis caching was introduced into the resolve path**. The results showed that caching significantly reduced request latency under higher loads by reducing the amount of work reaching PostgreSQL. For example, at 6,000 RPS, P99 decreased from **5.69 seconds without cache to 204.83 ms with cache**.

At this stage, however, the `url-service` itself reached a resource saturation point, limiting our ability to increase the RPS further in the current environment.

Therefore, the next step is to address the resource limitations of the `url-service`. This can be approached either by **vertical scaling**, by allocating more resources to the service, or by **horizontal scaling**, by running multiple instances of the `url-service` behind a load balancer.

After addressing the service-level resource limitation, the performance tests will be repeated at higher RPS levels. The same workloads will then be tested **both without cache and with cache** again.

This will allow us to determine how the system scales after increasing `url-service` capacity and provide a more meaningful comparison of the performance impact of Redis caching at higher throughput levels.
