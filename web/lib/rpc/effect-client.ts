import {
  type CallOptions,
  Code,
  type ResponseStream,
  type RpcServiceDescriptor,
  TrevRpcError,
} from "@trevrpc/trevrpc-js";
import { Effect, Stream } from "effect";

export type EffectClient<Client> = {
  [P in keyof Client]: Client[P] extends (
    request: infer I,
    options?: CallOptions,
  ) => Promise<ResponseStream<infer O>>
    ? (request: I, options?: CallOptions) => Stream.Stream<O, TrevRpcError>
    : Client[P] extends (request: infer I, options?: CallOptions) => Promise<infer O>
      ? (request: I, options?: CallOptions) => Effect.Effect<O, TrevRpcError>
      : never;
};

type UnaryFn = (req: unknown, opts?: CallOptions) => Promise<unknown>;
type ServerStreamingFn = (req: unknown, opts?: CallOptions) => Promise<ResponseStream<unknown>>;

export function toTrevRpcError(error: unknown): TrevRpcError {
  if (error instanceof TrevRpcError) return error;
  const message = error instanceof Error ? error.message : String(error);
  return new TrevRpcError(Code.Unknown, message, undefined, { cause: error });
}

function liftUnary(fn: UnaryFn) {
  return (request: unknown, options?: CallOptions) =>
    Effect.tryPromise({
      try: (signal) => fn(request, { ...options, signal }),
      catch: toTrevRpcError,
    });
}

function liftServerStreaming(fn: ServerStreamingFn) {
  return (request: unknown, options?: CallOptions) =>
    Stream.unwrap(
      Effect.gen(function* () {
        const controller = new AbortController();
        yield* Effect.addFinalizer(() => Effect.sync(() => controller.abort()));
        const responses = yield* Effect.tryPromise({
          try: () => fn(request, { ...options, signal: controller.signal }),
          catch: toTrevRpcError,
        });
        return Stream.fromAsyncIterable(responses, toTrevRpcError);
      }),
    );
}

/** Wraps a generated client's unary and server-streaming methods in Effect. */
export function createEffectClient<Client extends object>(
  service: RpcServiceDescriptor,
  client: Client,
): EffectClient<Client> {
  const result: Record<string, unknown> = {};
  const raw = client as unknown as Record<string, UnaryFn & ServerStreamingFn>;
  for (const [name, method] of Object.entries(service.methods)) {
    const fn = raw[name].bind(client);
    if (method.kind === "unary") {
      result[name] = liftUnary(fn);
    } else if (method.kind === "serverStreaming") {
      result[name] = liftServerStreaming(fn);
    }
  }
  return result as EffectClient<Client>;
}
