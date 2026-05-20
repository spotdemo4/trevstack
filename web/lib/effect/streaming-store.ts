import { Effect, Fiber, Stream } from "effect";
import { createEffect, createSignal, on, onCleanup, type Accessor } from "solid-js";
import { createStore } from "solid-js/store";

export type StreamingStore<Item> = {
  items: Item[];
  loading: Accessor<boolean>;
};

export function createStreamingStore<Req, Resp, Item, E>(
  request: Accessor<Req>,
  stream: (req: Req) => Stream.Stream<Resp, E>,
  map: (resp: Resp) => Item,
): StreamingStore<Item> {
  const [items, setItems] = createStore<Item[]>([]);
  const [loading, setLoading] = createSignal(true);
  let fiber: Fiber.RuntimeFiber<void, E> | null = null;
  let runId = 0;

  createEffect(
    on(request, (req) => {
      const currentRunId = ++runId;
      if (fiber) Effect.runFork(Fiber.interrupt(fiber));

      setLoading(true);
      setItems([]);

      let pending: Item[] = [];
      let rafId = 0;
      const flush = () => {
        rafId = 0;
        if (currentRunId !== runId) {
          pending = [];
          return;
        }
        if (!pending.length) return;
        const batch = pending;
        pending = [];
        setItems((prev) => [...prev, ...batch]);
      };

      onCleanup(() => {
        if (rafId) cancelAnimationFrame(rafId);
      });

      const program = stream(req).pipe(
        Stream.runForEach((resp) =>
          Effect.sync(() => {
            if (currentRunId !== runId) return;
            pending.push(map(resp));
            if (!rafId) rafId = requestAnimationFrame(flush);
          }),
        ),
        Effect.tap(() => Effect.sync(flush)),
        Effect.tapErrorCause((cause) => Effect.sync(() => console.error(cause))),
        Effect.ensuring(
          Effect.sync(() => {
            flush();
            if (currentRunId === runId) setLoading(false);
          }),
        ),
      );

      fiber = Effect.runFork(program);
    }),
  );

  onCleanup(() => {
    runId += 1;
    if (fiber) Effect.runFork(Fiber.interrupt(fiber));
  });

  return { items, loading };
}
