import { Effect, Fiber, Stream } from "effect";
import { type Accessor, createSignal, onSettled } from "solid-js";

/** `live` while the stream runs, `ended` once it completes, and `failed` once it fails. */
export type LiveStreamStatus = "live" | "ended" | "failed";

export type LiveStream = {
  status: Accessor<LiveStreamStatus>;
  /** The failure of a `failed` stream. */
  error: Accessor<unknown>;
  /** Whether a `failed` stream will restart by itself. */
  retrying: Accessor<boolean>;
  /** Stops the stream, if it is running, and starts it again. */
  restart: () => void;
};

export type LiveStreamOptions = {
  /** How long to wait before restarting a failed stream. 3 seconds by default; 0 never restarts. */
  retryDelayMs?: number;
};

/**
 * Runs `stream` from when the component settles until it is cleaned up, passing each value to
 * `onItem` as it arrives, such as the responses of a server-streaming or bidirectional call that
 * update a page live. A stream that fails restarts after a delay. Unlike `createStreamingStore`,
 * it keeps nothing, so it suits streams that never end.
 */
export function createLiveStream<A, E>(
  stream: () => Stream.Stream<A, E>,
  onItem: (item: A) => void,
  options: LiveStreamOptions = {},
): LiveStream {
  const retryDelayMs = options.retryDelayMs ?? 3000;
  const [status, setStatus] = createSignal<LiveStreamStatus>("live");
  const [error, setError] = createSignal<unknown>();
  const [retrying, setRetrying] = createSignal(false);
  let fiber: Fiber.Fiber<void> | undefined;
  let retryTimer: ReturnType<typeof setTimeout> | undefined;

  const stop = () => {
    clearTimeout(retryTimer);
    retryTimer = undefined;
    setRetrying(false);
    if (fiber) Effect.runFork(Fiber.interrupt(fiber));
    fiber = undefined;
  };

  const start = () => {
    stop();
    setError(undefined);
    setStatus("live");
    fiber = Effect.runFork(
      stream().pipe(
        Stream.runForEach((item) => Effect.sync(() => onItem(item))),
        Effect.match({
          onSuccess: () => {
            setStatus("ended");
          },
          onFailure: (err) => {
            console.error(err);
            setError(() => err);
            setStatus("failed");
            if (retryDelayMs > 0) {
              setRetrying(true);
              retryTimer = setTimeout(start, retryDelayMs);
            }
          },
        }),
      ),
    );
  };

  onSettled(() => {
    start();
    return stop;
  });

  return { status, error, retrying, restart: start };
}
