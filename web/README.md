# web

The TrevStack web app is a [SolidJS](https://www.solidjs.com/) single-page app that talks to the
server's [TrevRPC](https://trev.zip/llc/TrevRPC) API. It is built with [Vite](https://vite.dev/), styled with
[Tailwind CSS](https://tailwindcss.com/), and installable as a progressive web app.

The production build is embedded in and served by the [server](../server/README.md), so there is
nothing to deploy separately.

## pages

| path       | description                                             |
| ---------- | ------------------------------------------------------- |
| `/auth`    | sign in and sign up                                     |
| `/`        | add a named number                                      |
| `/numbers` | filter and browse recorded numbers as they stream in    |
| `/metrics` | summary, time series, distribution, and top name charts |

Every page except `/auth` requires a session; signed-out visitors are redirected to `/auth` and
returned to the page they asked for after signing in. RPCs that fail with `permission_denied` show
`/403`.

## api

RPCs go to the server's `/trevrpc` endpoint over one channel the app shares, which starts connecting
when the app loads: WebTransport, or QMux over WebSocket when WebTransport is unavailable or UDP is
blocked. Calls made while it connects wait for it.
Clients for each service are in [`lib/rpc`](lib/rpc) and wrap calls in
[Effect](https://effect.website/) with [`@trevrpc/effect`](https://www.npmjs.com/package/@trevrpc/effect)
for typed errors and cancellation. The clients check each request against the
[`buf.validate`](https://buf.build/bufbuild/protovalidate) rules in its proto with
[`@trevrpc/protovalidate`](https://www.npmjs.com/package/@trevrpc/protovalidate) before sending it,
and forms check their values against the same rules with `protoSchema` from
[`lib/form/proto-schema.ts`](lib/form/proto-schema.ts), so a mistake shows under its field at once
and the rules live in one place.

When the server uses its development certificate, the channel reads the certificate's hashes from
`/certificate-hashes` before each WebTransport attempt and trusts it by them.

`Login` returns a session token, which the app keeps in local storage with the username and expiry.
The channel sends it as connection metadata, which crosses each connection once instead of in every
request, and its interceptor handles the failures of calls that require a session. The app can tell from local
storage, without a request, whether to show signed-in pages. It signs out when the session expires
or when an RPC fails with `unauthenticated`, and syncs sign-in state across tabs. Logging out forgets
the token at once and asks the server, with `Logout`, to end the session.

## progressive web app

A service worker caches the app shell for offline use; `/trevrpc` and `/certificate-hashes` are never served from the cache.
The app checks for a new version every 15 minutes and whenever it comes back into view,
then offers the update in a toast instead of reloading, so unsaved input is never lost.

## theme

Colors come from semantic tokens in [`theme.css`](theme.css), such as `background`, `surface`,
`foreground-muted`, `line`, `primary`, and `danger`, which Tailwind exposes as utilities like
`bg-surface` and as `var(--color-surface)` in CSS. The defaults are [Catppuccin](https://catppuccin.com/)
Latte for the light theme and Mocha for the dark theme; change the values there to retheme the app.
The web app manifest and the colors shown before the stylesheet loads are read from the same file.

## development

Start the server in another terminal (see [server development](../server/README.md#development)),
then, inside the development shell, from this directory:

```sh
npm install
npm run dev
```

The app is served at http://localhost:3000 with hot reloading. It calls the server at
`https://localhost:8080/trevrpc`, set by `VITE_RPC_URL` in [`.env.development`](.env.development),
which admits the app's pages when built with the `dev` tag, and reads the hashes of the server's
development certificate through a proxy at `/certificate-hashes`. The development certificate is
trusted for WebTransport only, so use a browser with WebTransport, such as Chromium or Firefox;
the WebSocket fallback needs a certificate the browser trusts, passed with `TLS_CERT` and `TLS_KEY`.

| command         | description                                 |
| --------------- | ------------------------------------------- |
| `npm run dev`   | start the development server                |
| `npm run build` | build the production app into `dist/`       |
| `npm run serve` | preview the production build                |
| `npm test`      | run the [Vitest](https://vitest.dev/) tests |

Formatting and linting use [oxfmt](https://oxc.rs/docs/guide/usage/formatter) and
[oxlint](https://oxc.rs/docs/guide/usage/linter); both run in `nix flake check`.

### layout

| path        | description                                                  |
| ----------- | ------------------------------------------------------------ |
| `index.tsx` | entry point, routes, and session gating                      |
| `index.css` | Tailwind setup                                               |
| `theme.css` | theme colors                                                 |
| `routes/`   | one directory per page, with the page in `index.tsx`         |
| `layout/`   | navigation bar, user menu, theme switch, and network status  |
| `lib/`      | shared components, auth, TrevRPC clients, and Effect helpers |
| `gen/`      | code generated by `buf generate`; do not edit                |
| `public/`   | static files copied into the build as-is                     |

Import shared code with the `$lib/*` alias and generated code with `$gen/*`.
