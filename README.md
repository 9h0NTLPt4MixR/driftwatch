# driftwatch

A CLI tool that detects configuration drift between deployed services and their declared state in version control.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Point `driftwatch` at your version-controlled config directory and a running environment to compare against:

```bash
# Check for drift between declared state and production
driftwatch check --source ./config --env production

# Watch continuously and alert on drift
driftwatch watch --source ./config --env staging --interval 60s

# Output a diff report
driftwatch check --source ./config --env production --output diff
```

Example output:

```
[DRIFT DETECTED] service: api-gateway
  expected: replicas=3  got: replicas=2
  expected: image=app:v1.4.2  got: image=app:v1.3.9

[OK] service: auth-service
[OK] service: worker
```

---

## Configuration

`driftwatch` can be configured via a `.driftwatch.yaml` file in your project root:

```yaml
source: ./config
environment: production
interval: 30s
notify: slack
```

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

[MIT](LICENSE)