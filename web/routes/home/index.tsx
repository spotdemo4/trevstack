import { Card } from "$lib/card";
import { ChevronRight } from "$lib/icon";
import { type Component, For } from "solid-js";

type Demo = {
  href: string;
  title: string;
  kind: string;
  rpc: string;
  description: string;
};

const demos: Demo[] = [
  {
    href: "/forms",
    title: "Forms",
    kind: "Unary",
    rpc: "Send",
    description:
      "Send one cursor position from a form. The form checks the same buf.validate rules the server enforces, so mistakes show before the request is sent.",
  },
  {
    href: "/tables",
    title: "Tables",
    kind: "Server streaming",
    rpc: "List, ListStreams",
    description:
      "Search recorded positions and streams. The server streams each matching row, and the table renders them as they arrive.",
  },
  {
    href: "/metrics",
    title: "Metrics",
    kind: "Unary",
    rpc: "Summary, TimeSeries, Heatmap, TopUsers",
    description:
      "Aggregate counts and distances, activity over time, a heatmap of where cursors go, and the most active users.",
  },
  {
    href: "/server-streams",
    title: "Server streams",
    kind: "Server streaming",
    rpc: "Watch",
    description:
      "Watch every cursor move live, whether it comes from a browser, the command-line client, or a form.",
  },
  {
    href: "/client-streams",
    title: "Client streams",
    kind: "Client streaming",
    rpc: "Track",
    description:
      "Stream your cursor to the server, one message per move, then get a summary of the stream when you stop.",
  },
  {
    href: "/bidi-streams",
    title: "Bidi streams",
    kind: "Bidirectional streaming",
    rpc: "Share",
    description:
      "Share your cursor and see everyone else's at the same time, both ways over a single call.",
  },
];

const Home: Component = () => (
  <div class="mx-auto flex max-w-5xl flex-col gap-8 p-4 py-8 md:py-12">
    <header class="flex flex-col gap-3">
      <h1 class="text-3xl font-bold">TrevStack</h1>
      <p class="max-w-prose text-foreground-subtle">
        A full-stack template with a Go server, this SolidJS app, and a Rust command-line client,
        which all talk over{" "}
        <a
          href="https://trev.zip/llc/TrevRPC"
          target="_blank"
          rel="noopener noreferrer"
          class="text-link hover:underline"
        >
          TrevRPC
        </a>
        . Each page below demonstrates a kind of RPC by tracking cursor positions, which every
        signed-in user shares. The app keeps one connection to the server, over WebTransport or
        WebSocket, and every call shares it.
      </p>
    </header>

    <nav aria-label="Demos" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <For each={demos}>
        {(demo) => (
          <a
            href={demo.href}
            class="group rounded-xl focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <Card class="flex h-full flex-col gap-3 p-5 transition-colors group-hover:border-line-strong">
              <div class="flex items-center justify-between gap-2">
                <h2 class="text-lg font-semibold">{demo.title}</h2>
                <ChevronRight
                  size={18}
                  class="text-foreground-faint transition-transform group-hover:translate-x-0.5 group-hover:text-foreground"
                />
              </div>
              <div class="flex flex-wrap items-center gap-2 text-xs">
                <span class="rounded-full bg-fill-subtle px-2.5 py-0.5 font-semibold text-foreground-muted">
                  {demo.kind}
                </span>
                <code class="font-mono text-foreground-subtle">{demo.rpc}</code>
              </div>
              <p class="text-sm text-foreground-subtle">{demo.description}</p>
            </Card>
          </a>
        )}
      </For>
    </nav>

    <Card class="flex flex-col gap-2 p-5">
      <h2 class="font-semibold">From a terminal</h2>
      <p class="text-sm text-foreground-subtle">
        The command-line client calls the same RPCs. In a terminal it turns the window into a pad
        and reads your mouse; with piped input it reads positions as lines of{" "}
        <code class="font-mono">x y</code>.
      </p>
      <pre class="overflow-x-auto rounded-md bg-sunken p-3 font-mono text-xs text-foreground-muted">
        {["client send 0.5 0.5", "client watch", "client track", "client share"].join("\n")}
      </pre>
    </Card>
  </div>
);

export default Home;
