import type {
  DescMethodBiDiStreaming,
  DescMethodClientStreaming,
  DescMethodServerStreaming,
  DescMethodUnary,
  DescService,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { type CallOptions, type Client, ConnectError } from "@connectrpc/connect";
import { Effect, Stream } from "effect";

export type EffectClient<Desc extends DescService> = {
  [P in keyof Desc["method"]]: Desc["method"][P] extends DescMethodUnary<infer I, infer O>
    ? (
        request: MessageInitShape<I>,
        options?: CallOptions,
      ) => Effect.Effect<MessageShape<O>, ConnectError>
    : Desc["method"][P] extends DescMethodServerStreaming<infer I, infer O>
      ? (
          request: MessageInitShape<I>,
          options?: CallOptions,
        ) => Stream.Stream<MessageShape<O>, ConnectError>
      : Desc["method"][P] extends DescMethodClientStreaming<infer I, infer O>
        ? (
            request: AsyncIterable<MessageInitShape<I>>,
            options?: CallOptions,
          ) => Effect.Effect<MessageShape<O>, ConnectError>
        : Desc["method"][P] extends DescMethodBiDiStreaming<infer I, infer O>
          ? (
              request: AsyncIterable<MessageInitShape<I>>,
              options?: CallOptions,
            ) => Stream.Stream<MessageShape<O>, ConnectError>
          : never;
};

type PromiseFn = (req: unknown, opts?: CallOptions) => Promise<unknown>;
type IterableFn = (req: unknown, opts?: CallOptions) => AsyncIterable<unknown>;

function liftPromise(fn: PromiseFn) {
  return (request: unknown, options?: CallOptions) =>
    Effect.tryPromise({
      try: (signal) => fn(request, { ...options, signal }),
      catch: (err) => ConnectError.from(err),
    });
}

function liftIterable(fn: IterableFn) {
  return (request: unknown, options?: CallOptions) =>
    Stream.unwrapScoped(
      Effect.gen(function* () {
        const controller = new AbortController();
        yield* Effect.addFinalizer(() => Effect.sync(() => controller.abort()));
        const iter = fn(request, { ...options, signal: controller.signal });
        return Stream.fromAsyncIterable(iter, (err) => ConnectError.from(err));
      }),
    );
}

export function createEffectClient<Desc extends DescService>(
  service: Desc,
  client: Client<Desc>,
): EffectClient<Desc> {
  const result: Record<string, unknown> = {};
  const raw = client as unknown as Record<string, PromiseFn | IterableFn>;
  for (const method of service.methods) {
    const name = method.localName;
    const fn = raw[name];
    if (method.methodKind === "unary" || method.methodKind === "client_streaming") {
      result[name] = liftPromise(fn as PromiseFn);
    } else {
      result[name] = liftIterable(fn as IterableFn);
    }
  }
  return result as EffectClient<Desc>;
}
