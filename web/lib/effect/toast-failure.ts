import { toaster } from "@trev.zip/solid-toast";
import { TrevRpcError } from "@trevrpc/trevrpc-js";
import { Effect } from "effect";

export const errorDescription = (err: unknown) => {
  if (err instanceof TrevRpcError) {
    return err.statusMessage || err.message;
  }

  if (err instanceof Error) {
    return err.message;
  }

  if (typeof err === "string") {
    return err;
  }

  return "An unexpected error occurred.";
};

export const toastFailure =
  (title: string) =>
  <A, E>(effect: Effect.Effect<A, E>): Effect.Effect<A | undefined> =>
    effect.pipe(
      Effect.tapError((err) =>
        Effect.sync(() => {
          toaster.error({
            title,
            description: errorDescription(err),
          });
        }),
      ),
      Effect.orElseSucceed(() => undefined),
    );
