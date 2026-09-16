import { Effect } from "effect";
import { createMemo, type Accessor } from "solid-js";

export function createEffectResource<R, A, E>(
  source: Accessor<R>,
  fetcher: (req: R) => Effect.Effect<A, E>,
): Accessor<A> {
  return createMemo(async () => {
    const req = source();
    return Effect.runPromise(fetcher(req));
  });
}
