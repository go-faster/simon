# simon

Observability load testing and simulation, under construction.

Supported platforms: {linux, windows, darwin}-{amd64, arm64, **riscv64**}

## Observability
- [x] Logs
- [x] Metrics
- [x] Traces
- [ ] Context propagation
- [ ] Profiling
- [ ] Health checks
  - [ ] Liveness
  - [ ] Readiness
  - [ ] Other probes

![hubble.png](_docs/hubble.png)

## Deployment
- [x] Kubernetes deployment
- [ ] Helm chart

## Run

### Docker
```console
docker run -i -t ghcr.io/go-faster/simon:latest
```

```console
{"level":"info","ts":1680181590.1318264,"logger":"metrics","caller":"app/metrics.go:286","msg":"No metrics exporter is configured by OTEL_METRICS_EXPORTER"}
{"level":"info","ts":1680181590.134922,"logger":"metrics","caller":"app/metrics.go:319","msg":"No traces exporter is configured by OTEL_TRACES_EXPORTER"}
{"level":"info","ts":1680181590.1362665,"logger":"metrics","caller":"app/metrics.go:356","msg":"Propagators configured","propagators":["tracecontext","baggage"]}
{"level":"info","ts":1680181590.1375117,"logger":"metrics","caller":"app/metrics.go:112","msg":"Registering pprof routes","routes":["profile","symbol","trace","goroutine","heap","threadcreate","block"]}
{"level":"info","ts":1680181590.139126,"logger":"metrics","caller":"app/metrics.go:379","msg":"Metrics initialized","otel.resource":"process.runtime.description=go version go1.20.2 linux/riscv64,process.runtime.name=go,process.runtime.version=go1.20.2,service.name=simon,service.namespace=go-faster,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.14.0","metrics.http.addr":"localhost:9464"}
{"level":"info","ts":1680181590.1406643,"logger":"metrics","caller":"app/metrics.go:62","msg":"Starting metrics server"}
{"level":"info","ts":1680181591.1405013,"caller":"simon/main.go:28","msg":"Hello, world!"}
{"level":"info","ts":1680181592.1402662,"caller":"simon/main.go:28","msg":"Hello, world!"}
```

### Pod

```console
kubectl -n sandbox apply -f _deploy/
```

## Environment variables


See [General SDK Configuration][general-sdk] for OpenTelemetry.

[general-sdk]: https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/


| Name                              | Description                                                                 | Example                 | Default                                                      |
|-----------------------------------|-----------------------------------------------------------------------------|-------------------------|--------------------------------------------------------------|
| `METRICS_ADDR`                    | Address with metrics and pprof                                              | `localhost:9464`        | To prometheus addr                                           |
| `PPROF_ROUTES`                    | List of enabled pprof routes                                                | `cmdline,profile`       | profile, symbol, trace, goroutine, heap, threadcreate, block |
| `OTEL_LOG_LEVEL`                  | Log level                                                                   | `debug`                 | `info`                                                       |
| `OTEL_EXPORTER_PROMETHEUS_HOST`   | Host of prometheus addr                                                     | `0.0.0.0`               | `localhost`                                                  |
| `OTEL_EXPORTER_PROMETHEUS_PORT`   | Port of prometheus addr                                                     | `9090`                  | `9464`                                                       |
| `OTEL_METRICS_EXPORTER`           | Metrics exporter to use                                                     | `prometheus`            | `none`                                                       |
| `OTEL_TRACES_EXPORTER`            | Traces exporter to use                                                      | `jaeger`                | `none`                                                       |
| `OTEL_EXPORTER_JAEGER_AGENT_HOST` | Jaeger exporter host                                                        | `jaeger.svc.local`      | `localhost`                                                  |
| `OTEL_EXPORTER_JAEGER_AGENT_PORT` | Jaeger exporter port                                                        | `6831`                  | `6831`                                                       |
| `OTEL_SERVICE_NAME`               | OTEL Service name                                                           | `app`                   | `unknown_service`                                            |
| `OTEL_RESOURCE_ATTRIBUTES`        | OTEL Resource attributes                                                    | `service.namespace=pfm` |                                                              |
| `OTEL_PROPAGATORS`                | OTEL Propagators (only tracecontext and baggage supported, none to disable) | `none`                  | `tracecontext,baggage`                                       |


| Environment variable                                                     | Option                        | Default value                                            |
|--------------------------------------------------------------------------|-------------------------------|----------------------------------------------------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`       | `WithEndpoint` `WithInsecure` | `https://localhost:4317` or `https://localhost:4318`[^1] |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` `OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE` | `WithTLSClientConfig`         |                                                          |
| `OTEL_EXPORTER_OTLP_HEADERS` `OTEL_EXPORTER_OTLP_TRACES_HEADERS`         | `WithHeaders`                 |                                                          |
| `OTEL_EXPORTER_OTLP_COMPRESSION` `OTEL_EXPORTER_OTLP_TRACES_COMPRESSION` | `WithCompression`             |                                                          |
| `OTEL_EXPORTER_OTLP_TIMEOUT` `OTEL_EXPORTER_OTLP_TRACES_TIMEOUT`         | `WithTimeout`                 | `10s`                                                    |


### `METRICS_ADDR`

| Value       | `METRICS_ADDR`                 |
|-------------|--------------------------------|
| Default     | `localhost:9464`               |
| Description | Address with metrics and pprof |
| Example     | `0.0.0.0:9464`                 |

### `PPROF_ROUTES`

| Value       | `PPROF_ROUTES`                                               |
|-------------|--------------------------------------------------------------|
| Default     | profile, symbol, trace, goroutine, heap, threadcreate, block |
| Description | List of enabled pprof routes                                 |
