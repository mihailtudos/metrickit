# Live agent-server metrics collector

*MetricKit* - a distributed monitoring solution for Go applications. Deploy lightweight agents to collect runtime and custom metrics from multiple systems, with centralized analysis, real-time alerting, and flexible reporting. Built for production environments requiring comprehensive observability with minimal overhead.

## Motivation

Distributed systems demand continuous monitoring to maintain reliability and performance, yet many existing solutions are either too heavyweight for smaller deployments or lack the granular control needed for Go-specific runtime metrics.

MetricKit was born from the need for:

*🎯 Proactive System Health*

Real-time visibility into critical system parameters before they become incidents. Understanding memory pressure, GC behavior, and resource utilization patterns helps prevent cascading failures in production environments.

*⚡ Lightweight Monitoring*

Many monitoring solutions consume significant resources themselves. MetricKit prioritizes minimal overhead while providing comprehensive insights - essential for resource-constrained environments or cost-sensitive deployments.

*🔧 Go-Native Observability*

Purpose-built for Go applications with deep runtime introspection. Track goroutine leaks, memory allocation patterns, GC pressure, and other Go-specific metrics that generic monitoring tools often miss or handle poorly.

*🎛️ Operational Control*

Fine-grained control over what gets monitored, how often, and when alerts fire. Different environments (development, staging, production) have different sensitivity requirements - MetricKit adapts to your operational needs rather than forcing you to adapt to the tool.

*📊 Learning Through Building*

Understanding monitoring systems deeply by building one from scratch. This hands-on approach reveals the complexities of distributed metrics collection, data aggregation, and alerting logic that using existing tools often abstracts away.

## Quick Start

To run the development version you will do the following:

1. Clone the repository:

```bash
    git clone https://github.com/mihailtudos/metrickit.git
```

2. Run the DB migration:

Before running the application, ensure you a running version Postgres DB:

```bash
    docker-compose up -d db
```

3. Install dependencies:

```shell
    go mod download
```

4. Run the server:

In order to run the server you will need to set the configuration value via the inline environment variable, flags, or configuration json files:

The priority order is:
 - Inline environment variable
 - Flags
 - Configuration file
 - Defaults

```shell
    go run ./cmd/server/. \
		-crypto-key="./private.pem" \
		-d="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" $(ARGS)
```

5. Run the agent:

```shell
    go run ./cmd/agent/. $(ARGS) \ 
    -crypto-key=public.pem \ 
    -address=localhost:50051
```

The agent can be also run with GRPC mode just by setting the `-grpc-addr` flag to the server address.

Usage
Contributing
