import { Effect } from "effect";
import { createResource, type ResourceReturn, type ResourceSource } from "solid-js";

export function createEffectResource<R, A, E>(
  source: ResourceSource<R>,
  fetcher: (req: R) => Effect.Effect<A, E>,
): ResourceReturn<A> {
  return createResource<A, R>(source, (req) => Effect.runPromise(fetcher(req)));
}

export const logFailure =
  (label: string) =>
  <A, E>(effect: Effect.Effect<A, E>): Effect.Effect<A | undefined> =>
    effect.pipe(
      Effect.tapError((err) => Effect.sync(() => console.error(`${label} failed`, err))),
      Effect.orElseSucceed(() => undefined),
    );
