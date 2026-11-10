import type { StreamSummary } from "$gen/cursor/v1/cursor_pb";
import type { TrackRequestSchema } from "$gen/cursor/v1/track_pb";
import { session } from "$lib/auth";
import { Card } from "$lib/card";
import { createCursorStore, CursorPad, type PadPoint } from "$lib/cursor-pad";
import { createPushStream, errorDescription, type PushStream } from "$lib/effect";
import { formatDuration } from "$lib/format";
import { PageHeader } from "$lib/page-header";
import { CursorClient } from "$lib/rpc";
import { toaster } from "$lib/toast";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { Effect, Fiber } from "effect";
import { type Component, createSignal, For, onSettled, Show } from "solid-js";

type TrackState = "idle" | "tracking" | "stopping";

/** The key of your own cursor. */
const OWN_CURSOR = "own";

const ClientStreams: Component = () => {
  const cursors = createCursorStore();
  const [state, setState] = createSignal<TrackState>("idle");
  const [sent, setSent] = createSignal(0);
  // Undefined until a stream ends; null when it ended without positions.
  const [summary, setSummary] = createSignal<StreamSummary | null>();
  let requests: PushStream<MessageInitShape<typeof TrackRequestSchema>> | undefined;
  let fiber: Fiber.Fiber<void> | undefined;

  const start = () => {
    const stream = createPushStream<MessageInitShape<typeof TrackRequestSchema>>();
    requests = stream;
    setSent(0);
    setState("tracking");
    fiber = Effect.runFork(
      CursorClient.track(stream.stream).pipe(
        Effect.match({
          onSuccess: (response) => {
            setSummary(() => response.stream ?? null);
            setState("idle");
          },
          onFailure: (err) => {
            requests = undefined;
            setState("idle");
            toaster.error({ title: "Tracking failed", description: errorDescription(err) });
          },
        }),
      ),
    );
  };

  const stop = () => {
    // Ending the request stream ends the call, and the server answers with its summary.
    requests?.end();
    requests = undefined;
    setState("stopping");
  };

  onSettled(() => () => {
    if (fiber) Effect.runFork(Fiber.interrupt(fiber));
  });

  const move = (point: PadPoint) => {
    if (requests) {
      requests.push(point);
      setSent((count) => count + 1);
    }
    cursors.move(OWN_CURSOR, session.claims()?.sub ?? "you", point);
  };

  const stats = (summary: StreamSummary) => [
    { label: "Stream", value: summary.id.toString() },
    { label: "Positions", value: summary.count.toLocaleString() },
    { label: "Distance", value: summary.distance.toFixed(2) },
    { label: "Duration", value: formatDuration(summary.start, summary.end) },
  ];

  return (
    <div class="mx-auto flex max-w-5xl flex-col gap-4 p-4">
      <PageHeader
        title="Client streams"
        kind="Client streaming"
        rpc="rpc Track(stream TrackRequest) returns (TrackResponse)"
      >
        Start tracking, and every move over the pad becomes a message in one stream to the server,
        which records each as it arrives. Stop, and the server answers once with a summary.
      </PageHeader>

      <div class="flex min-h-8 flex-wrap items-center gap-3">
        <button
          type="button"
          disabled={state() === "stopping"}
          class="inline-flex h-9.5 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:cursor-pointer hover:bg-primary-hover disabled:cursor-not-allowed disabled:bg-fill disabled:text-foreground-faint"
          onClick={() => (state() === "tracking" ? stop() : start())}
        >
          {state() === "idle" ? "Start tracking" : "Stop tracking"}
        </button>
        <Show when={state() !== "idle"}>
          <span role="status" class="text-sm text-foreground-muted tabular-nums">
            {sent().toLocaleString()} sent
          </span>
        </Show>
      </div>

      <CursorPad
        label="Tracking pad"
        cursors={cursors.cursors}
        onMove={move}
        onLeave={() => cursors.remove(OWN_CURSOR)}
        class="cursor-none"
      />

      <Show when={summary() !== undefined}>
        <Card class="p-4">
          <h2 class="mb-3 font-semibold">Last stream</h2>
          <Show
            when={summary()}
            fallback={<p class="text-sm text-foreground-subtle">No positions were sent.</p>}
          >
            {(summary) => (
              <dl class="grid grid-cols-2 gap-3 sm:grid-cols-4">
                <For each={stats(summary())}>
                  {(stat) => (
                    <div>
                      <dt class="text-xs tracking-wide text-foreground-subtle uppercase">
                        {stat.label}
                      </dt>
                      <dd class="mt-1 font-mono text-xl tabular-nums">{stat.value}</dd>
                    </div>
                  )}
                </For>
              </dl>
            )}
          </Show>
        </Card>
      </Show>
    </div>
  );
};

export default ClientStreams;
