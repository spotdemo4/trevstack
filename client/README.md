# client

The TrevStack client is a Rust command-line tool for the server's cursor service. It sends cursor
positions, streams your mouse position over the terminal, and shows everyone else's cursors as they
move, with one subcommand for each kind of RPC. It talks to the server with
[TrevRPC](https://trev.zip/llc/TrevRPC) over QUIC, and falls back to QMux over TLS when UDP is
blocked.

## usage

```sh
export STACK_USERNAME=demo_user STACK_PASSWORD=example-password

client send 0.25 0.75   # record one position, print its stream ID
client watch            # show every cursor as it moves
client track            # stream your mouse position, then print a summary
client share            # stream your mouse position and show everyone else's
```

### docker

```sh
docker run trev.zip/template/stack/client:latest
```

### nix

```sh
nix run git+https://trev.zip/template/stack#client
```

### download

https://trev.zip/template/stack/releases

### subcommands

Positions are fractions of the pad's width and height, from `0` to `1`, measured from its top-left
corner, so a position shows in the same place on every client whatever its size.

| subcommand | RPC                              | description                                                 |
| ---------- | -------------------------------- | ----------------------------------------------------------- |
| `send X Y` | `Send`, unary                    | record the position `X Y` and print its stream ID           |
| `watch`    | `Watch`, server streaming        | show every user's cursor as it moves                        |
| `track`    | `Track`, client streaming        | stream your mouse position, then print the stream's summary |
| `share`    | `Share`, bidirectional streaming | stream your mouse position while showing every other cursor |

Run in a terminal, `watch`, `track` and `share` take over the terminal as a pad: move the mouse over
it to send your position, and see other streams' cursors as coloured markers with their usernames.
The bottom line shows what the client is doing. Press `q`, `Esc` or `Ctrl-C` to stop. Cursors
disappear when their stream ends, or after 10 seconds without moving. The pad works in terminals with
mouse support on Linux, macOS and Windows, including under Wayland, since it reads the mouse from the
terminal rather than the desktop.

When stdin or stdout is not a terminal, they read and write lines instead, so they can be scripted:

- `track` and `share` send one position for each line of stdin, written as `x y`, and end their
  request stream at the end of the input. Blank lines are skipped; any other line that is not a
  position is an error.
- `watch` and `share` print one line for each event:
  `moved <stream_id> <username> <source> <x> <y>` when a cursor moves, where `source` is `send`,
  `track` or `share`, and `ended <stream_id> <username>` when a stream ends.
- `track` prints the summary of its stream once the server answers, as
  `<stream_id> <count> <distance>`, or `0 0 0.0000` when it sent no positions.

```sh
printf '0.1 0.1\n0.5 0.5\n0.9 0.9\n' | client track   # 12 3 1.1314
client watch | grep moved
```

`watch` runs until the server ends the stream or `Ctrl-C`. `share` stops reading stdin at `Ctrl-C`,
and exits once the server ends the call.

### options

| flag                  | environment      | default                  | description                                   |
| --------------------- | ---------------- | ------------------------ | --------------------------------------------- |
| `--url URL`           | `STACK_URL`      | `https://localhost:8080` | server URL                                    |
| `--ca-cert FILE`      | `STACK_CA_CERT`  |                          | PEM certificates to trust besides the bundled |
| `--username USERNAME` | `STACK_USERNAME` |                          | **required**, account username                |
| `--password PASSWORD` | `STACK_PASSWORD` |                          | **required**, account password                |
| `-h, --help`          |                  |                          | print help                                    |
| `-V, --version`       |                  |                          | print version                                 |

Options may come before or after the subcommand. Flags take precedence over environment variables.
Prefer `STACK_PASSWORD` over `--password` to keep the password out of the process list and shell
history.

The URL must use `https` and name only a host and an optional port (`443` by default), without a
path, credentials, a query string, or a fragment. The server's certificate is verified against the
bundled [webpki roots](https://github.com/rustls/webpki-roots) and the certificates in `--ca-cert`.
To trust a server's development certificate, save it with
`openssl s_client -connect localhost:8080 </dev/null | openssl x509 > server.pem` and pass
`--ca-cert server.pem`; save it again when the server replaces it, every 6 days.

### authentication

The account must already exist; create one from the web app or with `AuthService.Signup`. The
client calls `AuthService.Login` and sends the returned token as `authorization: Bearer <token>`
metadata on each cursor call, then calls `AuthService.Logout` to end the session, since the token is
not stored between runs. It logs out whether the command succeeds, fails, or is stopped.

### exit status

The client exits `0` on success. On failure it prints an error to stderr and exits `1`, or `2` when
the arguments cannot be parsed. Connecting times out after 5 seconds, and logging in, logging out and
`send` after 10 seconds each; the streaming subcommands run for as long as they are used.

## development

From the repository root, run the client with:

```sh
STACK_USERNAME=demo_user STACK_PASSWORD=example-password \
  nix run .#client -- watch
```

or, inside the development shell, from this directory:

```sh
STACK_USERNAME=demo_user STACK_PASSWORD=example-password \
  cargo run -- watch
cargo test
cargo clippy --all-targets
```

### layout

| path            | description                                                       |
| --------------- | ----------------------------------------------------------------- |
| `src/main.rs`   | entry point                                                       |
| `src/lib.rs`    | runs a subcommand, on the pad or over stdin and stdout            |
| `src/cli.rs`    | argument and environment parsing                                  |
| `src/client.rs` | URL validation, connecting, login and logout, and the `Send` call |
| `src/stream.rs` | the `Watch`, `Track` and `Share` calls                            |
| `src/pad.rs`    | the terminal pad                                                  |
| `src/event.rs`  | cursor events, and the lines the client reads and prints          |
| `src/proto.rs`  | module wrapper for the generated code                             |
| `src/gen/`      | code generated by `buf generate`; do not edit                     |
