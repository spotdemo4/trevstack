import { toaster } from "$lib/toast/toaster";
import { ConnectError } from "@connectrpc/connect";
import { Effect } from "effect";

const errorDescription = (err: unknown) => {
  if (err instanceof ConnectError) {
    return err.rawMessage || err.message;
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
