# hivecheck

> CLI tool for running structured health checks against microservices with configurable thresholds and alerting hooks.

---

## Installation

```bash
go install github.com/yourusername/hivecheck@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/hivecheck.git
cd hivecheck
go build -o hivecheck .
```

---

## Usage

Define your services in a `hivecheck.yaml` config file:

```yaml
services:
  - name: payments-api
    url: https://payments.internal/health
    interval: 30s
    thresholds:
      latency_ms: 500
      status_code: 200
    alerts:
      webhook: https://hooks.slack.com/your-webhook-url
```

Then run:

```bash
hivecheck run --config hivecheck.yaml
```

**Additional commands:**

```bash
hivecheck run --config hivecheck.yaml   # Start health check loop
hivecheck check --service payments-api  # Run a one-time check
hivecheck report                        # Print last known status for all services
hivecheck --help                        # Show all available commands
```

---

## Features

- Structured health checks with configurable polling intervals
- Latency and status code threshold enforcement
- Webhook alerting hooks (Slack, PagerDuty, custom endpoints)
- Simple YAML-based configuration
- Human-readable terminal output with JSON export support

---

## License

[MIT](LICENSE)