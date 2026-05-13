import { Effect, Fiber, Stream } from "effect";
import { type Accessor, createEffect, on, onCleanup } from "solid-js";
import { createStore } from "solid-js/store";

export function createStreamingStore<Req, Resp, Item, E>(
  request: Accessor<Req>,
  stream: (req: Req) => Stream.Stream<Resp, E>,
  map: (resp: Resp) => Item,
): Item[] {
  const [items, setItems] = createStore<Item[]>([]);
  let fiber: Fiber.RuntimeFiber<void, E> | null = null;

  createEffect(
    on(request, (req) => {
      if (fiber) Effect.runFork(Fiber.interrupt(fiber));

      setItems([]);

      let pending: Item[] = [];
      let rafId = 0;
      const flush = () => {
        rafId = 0;
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
            pending.push(map(resp));
            if (!rafId) rafId = requestAnimationFrame(flush);
          }),
        ),
        Effect.tap(() => Effect.sync(flush)),
        Effect.tapErrorCause((cause) => Effect.sync(() => console.error(cause))),
      );

      fiber = Effect.runFork(program);
    }),
  );

  onCleanup(() => {
    if (fiber) Effect.runFork(Fiber.interrupt(fiber));
  });

  return items;
}
