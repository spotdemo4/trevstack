# server

The TrevStack server is a single Go binary that serves the ConnectRPC API and the web app. Data is
stored in an embedded SQLite database.

## usage

```sh
JWT_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  server --port 8080
```

### options

| flag                | environment | default | description                                          |
| ------------------- | ----------- | ------- | ---------------------------------------------------- |
| `--port PORT`       | `PORT`      | `8080`  | port to listen on                                    |
| `--log-level LEVEL` | `LOG_LEVEL` | `info`  | minimum log level (`debug`, `info`, `warn`, `error`) |

Flags take precedence over environment variables.

### environment

| variable              | default | description                                                                              |
| --------------------- | ------- | ---------------------------------------------------------------------------------------- |
| `JWT_SECRET`          |         | **required**, secret used to sign session tokens; at least 32 bytes                      |
| `AUTH_COOKIE_SECURE`  | `false` | mark the session cookie `Secure`; enable when served over HTTPS                          |
| `TRUSTED_PROXY_CIDRS` |         | comma-separated proxy CIDRs (e.g. `10.0.0.0/8,::1/128`) allowed to set `X-Forwarded-For` |
| `OTEL_*`              |         | OpenTelemetry SDK settings, see [opentelemetry](#opentelemetry)                          |

## routes

| path     | description                                                              |
| -------- | ------------------------------------------------------------------------ |
| `/grpc/` | ConnectRPC API, accepting the Connect, gRPC, and gRPC-Web protocols      |
| `/docs/` | API reference generated from the proto definitions                       |
| `/`      | web app; unknown paths fall back to `index.html` for client-side routing |

The server speaks HTTP/1.1 and unencrypted HTTP/2 (h2c), so gRPC clients can connect without TLS.
Terminate TLS at a reverse proxy in production. CORS allows every origin.

### services

- `auth.v1.AuthService`: `Signup`, `Login`, and `Logout`. These do not require a session.
- `number.v1.NumberService`: `Add`, `List`, `Summary`, `TimeSeries`, `Distribution`, and `TopNames`.
  These require a session.

The proto definitions in [`../proto`](../proto) are the source of truth for every request and response.

## authentication

`Login` returns an HS256 JWT signed with `JWT_SECRET` and sets it in the `stack_session` cookie
(`HttpOnly`, `SameSite=Lax`, `Path=/`). Sessions last 24 hours. `Logout` clears the cookie, but does
not revoke tokens that were already issued, so rotating `JWT_SECRET` is the only way to end every
session.

Non-browser clients can send the token themselves in a `Cookie: stack_session=<jwt>` header.

### rate limiting

`Login` is limited to 5 requests per minute and `Signup` to 3 requests per hour for each client IP.
Rejected requests fail with `resource_exhausted` and a `Retry-After` header.

The client IP is the connection's peer address. When the peer is in `TRUSTED_PROXY_CIDRS`, the
server walks `X-Forwarded-For` from right to left and uses the first address that is not a trusted
proxy. Without `TRUSTED_PROXY_CIDRS`, `X-Forwarded-For` is ignored, so set it when running behind a
reverse proxy or every request will share the proxy's limit.

## database

SQLite is stored at `trevstack/trevstack.db` inside the user configuration directory:

| platform | path                                                                             |
| -------- | -------------------------------------------------------------------------------- |
| linux    | `$XDG_CONFIG_HOME/trevstack/trevstack.db`, or `~/.config/trevstack/trevstack.db` |
| macos    | `~/Library/Application Support/trevstack/trevstack.db`                           |
| windows  | `%AppData%\trevstack\trevstack.db`                                               |

The directory is created with `0700` permissions and the database with `0600`. Migrations in
[`database/migrations`](database/migrations) run automatically at startup. Mount a volume at the
configuration directory to persist data in a container.

## logging

Logs are written to stdout as JSON. Builds with the `dev` tag print human-readable logs instead.
RPCs are logged at `debug` and failed RPCs at `error`.

When the logs signal is enabled, the same records are also exported over OpenTelemetry.

## opentelemetry

The server exports traces, metrics, and logs using the standard
[OpenTelemetry environment variables](https://opentelemetry.io/docs/languages/sdk-configuration/).
Nothing is exported by default: each signal is enabled only when an exporter or endpoint is configured.

```sh
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
JWT_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  server
```

| variable                                            | description                                                                           |
| --------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `OTEL_EXPORTER_OTLP_ENDPOINT`                       | enable every signal and export it over OTLP to this endpoint                          |
| `OTEL_EXPORTER_OTLP_{TRACES,METRICS,LOGS}_ENDPOINT` | enable one signal and export it over OTLP to this endpoint                            |
| `OTEL_{TRACES,METRICS,LOGS}_EXPORTER`               | enable one signal with `otlp`, `console`, or `none`; metrics also accept `prometheus` |
| `OTEL_EXPORTER_OTLP_PROTOCOL`                       | `http/protobuf` (default) or `grpc`                                                   |
| `OTEL_EXPORTER_OTLP_HEADERS`                        | headers sent with every export, e.g. for authentication                               |
| `OTEL_SERVICE_NAME`                                 | service name (default: `trevstack-server`)                                            |
| `OTEL_RESOURCE_ATTRIBUTES`                          | extra resource attributes; can also override `service.version`                        |
| `OTEL_SDK_DISABLED`                                 | `true` disables all telemetry                                                         |

Any other variable supported by the OpenTelemetry Go SDK, such as `OTEL_TRACES_SAMPLER` or
`OTEL_METRIC_EXPORT_INTERVAL`, also applies. With `OTEL_METRICS_EXPORTER=prometheus`, metrics are
served for scraping at `localhost:9464/metrics`; change the address with
`OTEL_EXPORTER_PROMETHEUS_HOST` and `OTEL_EXPORTER_PROMETHEUS_PORT`.

### signals

- **traces**: a span for each RPC, with child spans for the SQL queries it runs. Queries are
  recorded as written, so bound parameters are never exported. Trace context sent by callers is
  linked to, rather than used as the parent of, the server's spans, since the API is public.
- **metrics**: RPC duration, message size, and request counts; database connection pool stats; and
  Go runtime metrics.
- **logs**: every log record at or above the configured log level.

Resources include the service name and version, host, process runtime, and SDK details.

## development

From the repository root, run the server with:

```sh
nix run .#server
```

or, inside the development shell, from this directory:

```sh
JWT_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  go run -tags dev .
```

The `dev` build tag enables human-readable logs and does not embed the web app, which is served by
its own development server instead; see [web development](../web/README.md#development).

With the tag, the API reference is served from `../docs/dist`; build it with `npm run build` in [`../docs`](../docs).

Without the tag, the build embeds the web app from `web/`, which the nix build copies from its package.
The API reference is embedded from `docs/` the same way.

Test both variants with:

```sh
go test ./...
go test -tags dev ./...
```

### layout

| path            | description                                          |
| --------------- | ---------------------------------------------------- |
| `main.go`       | startup, routing, and interceptor wiring             |
| `config.go`     | flag and environment parsing                         |
| `auth/`         | session tokens and cookies                           |
| `connect/`      | code generated by `buf generate`; do not edit        |
| `database/`     | SQLite connection and migrations                     |
| `handlers/`     | RPC service implementations and static file handlers |
| `interceptors/` | auth, CORS, logging, and rate limiting               |
| `logger/`       | `slog` setup and OpenTelemetry log bridge            |
| `telemetry/`    | OpenTelemetry provider setup                         |

SQL queries live next to the handlers that use them as `.sql` files and are embedded at build time.
