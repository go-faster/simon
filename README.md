# simon

Observability load testing and simulation, under construction.

## Observability
- [ ] Logs
- [ ] Metrics
- [ ] Traces
- [ ] Profiling

## Deployment
- [ ] Kubernetes deployment
- [ ] Helm chart

## Environment variables

| Name                              | Description                                                                 | Example                                  | Default                                                  |
|-----------------------------------|-----------------------------------------------------------------------------|------------------------------------------|----------------------------------------------------------|
| `METRICS_ADDR`                    | Address with metrics and pprof                                              | `localhost:9464`                         | To prometheus addr                                       |
| `GO_PPROF_ROUTES`                 | List of enabled pprof routes                                                | `cmdline,profile`                        | `profile,symbol,trace,goroutine,heap,threadcreate,block` |
| `OTEL_LOG_LEVEL`                  | Log level                                                                   |                                          | `info`                                                   |
| `OTEL_EXPORTER_PROMETHEUS_HOST`   | Host of prometheus addr                                                     | `0.0.0.0`                                | `localhost`                                              |
| `OTEL_EXPORTER_PROMETHEUS_PORT`   | Port of prometheus addr                                                     | `9090`                                   | `9464`                                                   |
| `OTEL_METRICS_EXPORTER`           | Metrics exporter to use                                                     | `prometheus`                             | `none`                                                   |
| `OTEL_TRACES_EXPORTER`            | Traces exporter to use                                                      | `jaeger`                                 | `none`                                                   |
| `OTEL_EXPORTER_JAEGER_AGENT_HOST` | Jaeger exporter host                                                        | `jaeger.svc.local`                       | `localhost`                                              |
| `OTEL_EXPORTER_JAEGER_AGENT_PORT` | Jaeger exporter port                                                        | `6831`                                   | `6831`                                                   |
| `OTEL_SERVICE_NAME`               | OTEL Service name                                                           | `app`                                    | `unknown_service`                                        |
| `OTEL_RESOURCE_ATTRIBUTES`        | OTEL Resource attributes                                                    | `service.name=app,service.namespace=pfm` |                                                          |
| `OTEL_PROPAGATORS`                | OTEL Propagators (only tracecontext and baggage supported, none to disable) | `none`                                   | `tracecontext,baggage`                                   |
