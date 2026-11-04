import { Effect, Fiber, Stream } from "effect";
import { createEffect, createSignal, createStore, type Accessor } from "solid-js";

export type StreamingStore<Item> = {
  items: Item[];
  loading: Accessor<boolean>;
};

/** How long to take responses before letting the browser render them. */
const FRAME_BUDGET_MS = 8;

/**
 * Returns to the event loop, so the browser can render and handle input. A
 * timer, unlike scheduler.yield(), whose continuation runs before rendering,
 * lets the next animation frame flush the responses taken so far.
 */
const yieldToBrowser = Effect.promise(
  (): Promise<void> => new Promise((resolve) => setTimeout(resolve, 0)),
);

export function createStreamingStore<Req, Resp, Item, E>(
  request: Accessor<Req>,
  stream: (req: Req) => Stream.Stream<Resp, E>,
  map: (resp: Resp) => Item,
): StreamingStore<Item> {
  const [items, setItems] = createStore<Item[]>([]);
  const [loading, setLoading] = createSignal(true);
  let fiber: Fiber.Fiber<void, E> | null = null;
  let runId = 0;

  createEffect(
    () => request(),
    (req) => {
      const currentRunId = ++runId;
      if (fiber) Effect.runFork(Fiber.interrupt(fiber));

      setLoading(true);
      setItems((draft) => {
        draft.length = 0;
      });

      let pending: Item[] = [];
      let rafId = 0;
      let sliceStart = performance.now();
      const flush = () => {
        rafId = 0;
        if (currentRunId !== runId) {
          pending = [];
          return;
        }
        if (!pending.length) return;
        const batch = pending;
        pending = [];
        setItems((draft) => {
          draft.push(...batch);
        });
      };

      const program = stream(req).pipe(
        Stream.runForEach((resp) =>
          Effect.suspend(() => {
            if (currentRunId !== runId) return Effect.void;
            pending.push(map(resp));
            if (!rafId) rafId = requestAnimationFrame(flush);

            // Responses that have already arrived are taken without returning
            // to the event loop, which would keep the browser from rendering
            // any of them until the stream ends.
            if (performance.now() - sliceStart < FRAME_BUDGET_MS) return Effect.void;
            return yieldToBrowser.pipe(
              Effect.tap(() => Effect.sync(() => (sliceStart = performance.now()))),
            );
          }),
        ),
        Effect.tap(() => Effect.sync(flush)),
        Effect.tapCause((cause) => Effect.sync(() => console.error(cause))),
        Effect.ensuring(
          Effect.sync(() => {
            flush();
            if (currentRunId === runId) setLoading(false);
          }),
        ),
      );

      fiber = Effect.runFork(program);

      return () => {
        runId += 1;
        if (rafId) cancelAnimationFrame(rafId);
        if (fiber) Effect.runFork(Fiber.interrupt(fiber));
      };
    },
  );

  return { items, loading };
}
