## TrevStack

This is a CRUD app to use as a template for starting projects

### Features

- **Communicate anywhere**. Define a [protocol buffer](https://protobuf.dev/), and [Connect](https://connectrpc.com/) generates type-safe code to facilitate communication between the server and any client (web, mobile, embedded, etc). The protocol buffers can contain annotations to validate fields on the client and server. For clients that cannot use Connect, an OpenAPI spec is also generated
- **Build anywhere**. The dev environment, testing and building is all declared in a single [Nix](https://nixos.org/) flake. Every developer and server can use the same environment
- **Deploy anywhere**. CI/CD is already set up using github actions. New versions are automatically released for every major platform, along with a docker image. The binaries created require zero run-time dependencies and are relatively small (this app is 26 MiB)
- Can be entirely self-hosted
- Authentication is rolled in, including API key, fingerprint & passkey
- Automatic database migration on startup
- Light & dark modes with the [catppuccin](https://catppuccin.com/palette/) color palette
- Really good at running as a [progressive web app](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)

While I personally prefer using a Svelte frontend and Go backend, feel free to swap them out with whatever you like, you'll be surprised how easy it is.

## Getting Started

1. [Install Nix](https://nixos.org/download/)
2. Run `nix develop`
3. Create a `server/.env` file that looks something like

```env
KEY=changeme
PORT=8080
URL=http://localhost:5173
DATABASE_URL=sqlite:/home/trev/.config/trevstack/sqlite.db
```

4. Run `treli` to start the server & client

It's that simple. If you're feeling fancy, install [direnv](https://direnv.net/) and the dev environment will load automatically.

### Useful Commands

- `bump [major | minor]`: bumps the current version up one. Defaults to "patch" (0.0.1 -> 0.0.2)

- `nix build [#trevstack-(GOOS)-(GOARCH)]`: builds the application. Defaults to building for your current platform, but can be built to many by specifying the GOOS and GOARCH values

- `nix flake check`: runs all validations

- `buf lint proto` & `buf generate`: lints and generates code from protocol buffers

- `sqlc vet` & `sqlc generate`: verifies and generates code from SQL files

- `dbmate new` & `dbmate up`: creates a new migration file and runs pending migrations

### Github Actions

To use github actions for CI/CD, you'll need to create a fine-grained personal access token for the repository with the permissions:

- Contents (read and write)
- Pull requests (read and write)

And change some settings for the repository:

- General -> Allow auto-merge: true
- Rules -> Rulesets -> New ruleset
  - Branch targeting criteria: Default
  - Branch rules
    - Require status checks to pass -> Add checks -> "check"
- Actions -> General -> Workflow permissions
  - Read and write permissions: true
  - Allow GitHub Actions to create and approve pull requests: true
- Secrets and variables -> Actions -> Repository secrets
  - PAT: (personal access token)

### Gitea Actions

To use gitea actions for CI/CD, you'll need to create an [API token](https://docs.gitea.com/development/api-usage) with the scopes:

- write:repository
- write:package

And change some settings for the repository:

- Repository -> Delete pull request branch after merge by default: true
- Branches -> Add New Rule
  - Protected Branch Name Pattern: main
  - Enable Status Check: true
  - Status check patterns: Check / check\*
- Actions -> Secrets
  - PAT: (API token)

## Components

### Client

- **svelte 5** [[docs](https://svelte.dev/docs/svelte)] UI framework
- **tailwind 4** [[docs](https://tailwindcss.com/)] CSS framework
- **bits ui** [[docs](https://bits-ui.com/docs/)] headless components
- [components](https://github.com/spotdemo4/trevstack/tree/main/client/src/lib/ui) from **shadcn-svelte** [[docs](https://www.shadcn-svelte.com/docs)] altered to work with tailwind 4, fit the [catppuccin](https://catppuccin.com/palette/) color palette, and use [shallow routing](https://svelte.dev/docs/kit/shallow-routing)
- **connect rpc** [[docs](https://connectrpc.com/docs/web/)] to communicating with the server
- **protovalidate-es** [[docs](https://github.com/bufbuild/protovalidate-es)], along with a function [coolforms](https://github.com/spotdemo4/trevstack/blob/main/client/src/lib/coolforms/) to emulate the library [sveltekit-superforms](https://superforms.rocks/)
- **simplewebauthn** [[docs](https://simplewebauthn.dev/docs/packages/browser)] for passkey authentication
- **scalar** [[docs](https://github.com/scalar/scalar)] for displaying openapi specs
- **tw-animate-css** [[docs](https://github.com/Wombosvideo/tw-animate-css)] to animate with just tailwind classes
- vite [[docs](https://vite.dev/)] for dev server and bundling
- eslint [[docs](https://eslint.org/)] for linting
- prettier [[docs](https://prettier.io/)] for formatting
  - prettier-plugin-svelte [[docs](https://github.com/sveltejs/prettier-plugin-svelte)]
  - prettier-plugin-tailwindcss [[docs](https://github.com/tailwindlabs/prettier-plugin-tailwindcss)]
  - prettier-plugin-sort-imports [[docs](https://github.com/IanVS/prettier-plugin-sort-imports)]

### Server

- **go** [[docs](https://go.dev/doc/)]
- **connect rpc** [[docs](https://connectrpc.com/docs/go/)] to serve gRPC & HTTP requests
- **protovalidate-go** [[docs](https://github.com/bufbuild/protovalidate-go)] for validating those requests
- **sqlc** [[docs](https://docs.sqlc.dev/en/latest/)] because writing the SQL yourself is better than an ORM
- **go-webauthn** [[docs](https://github.com/go-webauthn/webauthn)] because webauthn is hard
- **dbmate** [[docs](https://github.com/amacneil/dbmate)] for database migrations
- revive [[docs](https://github.com/mgechev/revive)] for linting

### Protocol Buffers / gRPC

- **buf** [[docs](https://buf.build/docs/)] CLI for linting & code generation
- **protovalidate** [[docs](https://buf.build/docs/protovalidate/)] provides annotations to validate proto messages & fields
- **protoc-gen-connect-openapi** [[docs](https://github.com/sudorandom/protoc-gen-connect-openapi)] generates openapi specs
- protoc-gen-go [[docs](https://pkg.go.dev/google.golang.org/protobuf)]
- protoc-gen-connect-go [[docs](https://connectrpc.com/docs/go)]
- protoc-gen-es [[docs](https://connectrpc.com/docs/web/)]
