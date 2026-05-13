import { type Accessor, createEffect, on, onCleanup } from "solid-js";
import { createStore } from "solid-js/store";

export function createStreamingStore<Req, Resp, Item>(
  request: Accessor<Req>,
  stream: (req: Req, opts: { signal: AbortSignal }) => AsyncIterable<Resp>,
  map: (resp: Resp) => Item,
): Item[] {
  const [items, setItems] = createStore<Item[]>([]);
  let abort = new AbortController();

  createEffect(
    on(request, async (req) => {
      abort.abort();
      abort = new AbortController();
      const signal = abort.signal;

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

      try {
        for await (const resp of stream(req, { signal })) {
          pending.push(map(resp));
          if (!rafId) rafId = requestAnimationFrame(flush);
        }
        flush();
      } catch (e) {
        if (signal.aborted) return;
        throw e;
      }
    }),
  );

  onCleanup(() => abort.abort());

  return items;
}
