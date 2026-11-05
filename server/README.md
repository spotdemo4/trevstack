# server

The TrevStack server is a single Go binary that serves the [TrevRPC](https://trev.zip/llc/TrevRPC)
API and the web app. Data is stored in an embedded SQLite database.

## usage

```sh
JWT_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  server --port 8080 --tls-cert cert.pem --tls-key key.pem
```

### docker

```sh
docker run -P \
  -e JWT_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  trev.zip/template/stack/server:latest
```

### nix

```sh
nix run git+https://trev.zip/template/stack#server
```

### download

https://trev.zip/template/stack/releases

### options

| flag                | environment | default   | description                                          |
| ------------------- | ----------- | --------- | ---------------------------------------------------- |
| `--port PORT`       | `PORT`      | `8080`    | UDP and TCP port to listen on                        |
| `--tls-cert FILE`   | `TLS_CERT`  | generated | PEM certificate chain, see [tls](#tls)               |
| `--tls-key FILE`    | `TLS_KEY`   | generated | PEM private key for `--tls-cert`                     |
| `--log-level LEVEL` | `LOG_LEVEL` | `info`    | minimum log level (`debug`, `info`, `warn`, `error`) |

Flags take precedence over environment variables.

### environment

| variable              | description                                                                                                |
| --------------------- | ---------------------------------------------------------------------------------------------------------- |
| `JWT_SECRET`          | **required**, secret used to sign session tokens; at least 32 bytes                                        |
| `ALLOWED_ORIGINS`     | comma-separated origins of other sites whose pages may call the API (e.g. `https://app.example.com`)       |
| `TRUSTED_PROXY_CIDRS` | comma-separated CIDRs of TCP proxies that send PROXY protocol headers, see [rate limiting](#rate-limiting) |
| `OTEL_*`              | OpenTelemetry SDK settings, see [opentelemetry](#opentelemetry)                                            |

## routes

| path                  | description                                                              |
| --------------------- | ------------------------------------------------------------------------ |
| `/trevrpc`            | TrevRPC browser sessions, over WebTransport or QMux over WebSocket       |
| `/certificate-hashes` | hashes of the development certificate, see [tls](#tls)                   |
| `/`                   | web app; unknown paths fall back to `index.html` for client-side routing |

The server listens on one UDP and one TCP port with the same number. Native TrevRPC clients connect
over QUIC on the UDP port and fall back to QMux over TLS on the TCP port when UDP is blocked.
Browsers connect to `/trevrpc` over WebTransport on the UDP port, or QMux over WebSocket on the TCP
port, and load the web app over HTTP/3, HTTP/2, or HTTP/1.1. Every connection uses TLS.

Browser sessions are admitted from pages served by the server itself and from `ALLOWED_ORIGINS`,
and in builds without the `prod` tag from the web development server at `http://localhost:3000`; the
`Origin` check keeps other sites from calling the API with a visitor's browser.

### services

- `auth.v1.AuthService`: `Signup` and `Login`. These do not require a session.
- `number.v1.NumberService`: `Add`, `List`, `Summary`, `TimeSeries`, `Distribution`, and `TopNames`.
  These require a session.

The proto definitions in [`../proto`](../proto) are the source of truth for every request and response,
and their comments describe the rules the server validates requests against.

## tls

Set `TLS_CERT` and `TLS_KEY` to a certificate that clients trust. Terminate TLS in the server rather
than at a reverse proxy: QUIC needs the server's own TLS, and so does QMux over TLS. A proxy in front
of the server must forward UDP and TCP without decrypting them.

Without a certificate, the server uses TrevRPC's
[development certificate](https://trev.zip/llc/TrevRPC/wiki/WebTransport#development-certificates):
a self-signed ECDSA certificate for `localhost`, `127.0.0.1`, and `::1` that it generates at startup
and replaces every 6 days. Browsers can trust a certificate like this for WebTransport by its hashes,
which the server publishes as JSON at `/certificate-hashes` (404 with a configured certificate) and
the web app reads, but not for anything else, so it is for development only.

## authentication

`Login` returns an HS256 JWT signed with `JWT_SECRET`. Clients send it as `authorization: Bearer <jwt>`
metadata on every call that requires a session. Calls with a missing, malformed, or expired token fail
with `unauthenticated`, and calls with a token that fails verification with `permission_denied`.
Sessions last 24 hours, and signing out only forgets the token on the client, so rotating
`JWT_SECRET` is the only way to end every session.

### rate limiting

`Login` is limited to 5 requests per minute and `Signup` to 3 requests per hour for each client IP.
Rejected requests fail with `resource_exhausted` and `retry-after` metadata holding the seconds to
wait.

The client IP is the source address of the connection. QUIC load balancers must preserve it. TCP
proxies in `TRUSTED_PROXY_CIDRS` must send a
[PROXY protocol](https://www.haproxy.org/download/3.0/doc/proxy-protocol.txt) v1 or v2 header with
every connection, whose source address is used instead; connections from them without a valid header
are closed, and connections from anywhere else are never read for a header. `X-Forwarded-For` is
not used, since a proxy that sets it would have to terminate TLS.

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
RPCs that pass authorization are logged at `debug`, and those that then fail at `error`. TrevRPC's
server diagnostics are logged as `trevrpc: <phase>` at the level TrevRPC recommends, such as `debug`
for failed TLS handshakes and `error` for handler panics.

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
- **metrics**: RPC duration and message sizes by method and status, from
  [oteltrevrpc](https://trev.zip/llc/TrevRPC/src/branch/main/trevrpc-go/oteltrevrpc); database
  connection pool stats; and Go runtime metrics.
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
  go run .
```

This development build writes human-readable logs and does not embed the web app, which is served by
its own development server instead; see [web development](../web/README.md#development). It also
admits that server's pages, at `http://localhost:3000` and `http://127.0.0.1:3000`, so they can call
the API.

The `prod` build tag, which the nix build uses, writes JSON logs, admits no development origins, and
embeds the web app from `web/`, which the nix build copies from its package.

Test both variants with:

```sh
go test ./...
go test -tags prod ./...
```

### layout

| path            | description                                          |
| --------------- | ---------------------------------------------------- |
| `main.go`       | startup, routing, and TrevRPC server setup           |
| `config.go`     | flag and environment parsing                         |
| `tls.go`        | certificate loading                                  |
| `auth/`         | session tokens and the authorizer that checks them   |
| `database/`     | SQLite connection and migrations                     |
| `gen/`          | code generated by `buf generate`; do not edit        |
| `handlers/`     | RPC service implementations and static file handlers |
| `interceptors/` | logging and rate limiting                            |
| `logger/`       | `slog` setup and OpenTelemetry log bridge            |
| `telemetry/`    | OpenTelemetry provider setup                         |
| `validation/`   | request field checks                                 |

SQL queries live next to the handlers that use them as `.sql` files and are embedded at build time.
