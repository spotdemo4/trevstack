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
import { Effect } from "effect";

export type EffectClient<Desc extends DescService> = {
  [P in keyof Desc["method"]]: Desc["method"][P] extends DescMethodUnary<infer I, infer O>
    ? (
        request: MessageInitShape<I>,
        options?: CallOptions,
      ) => Effect.Effect<MessageShape<O>, ConnectError>
    : Desc["method"][P] extends DescMethodServerStreaming<infer I, infer O>
      ? (request: MessageInitShape<I>, options?: CallOptions) => AsyncIterable<MessageShape<O>>
      : Desc["method"][P] extends DescMethodClientStreaming<infer I, infer O>
        ? (
            request: AsyncIterable<MessageInitShape<I>>,
            options?: CallOptions,
          ) => Effect.Effect<MessageShape<O>, ConnectError>
        : Desc["method"][P] extends DescMethodBiDiStreaming<infer I, infer O>
          ? (
              request: AsyncIterable<MessageInitShape<I>>,
              options?: CallOptions,
            ) => AsyncIterable<MessageShape<O>>
          : never;
};

export function createEffectClient<Desc extends DescService>(
  service: Desc,
  client: Client<Desc>,
): EffectClient<Desc> {
  const result: Record<string, unknown> = {};
  const promiseClient = client as unknown as Record<
    string,
    (req: unknown, opts?: CallOptions) => Promise<unknown> | AsyncIterable<unknown>
  >;
  for (const method of service.methods) {
    const name = method.localName;
    const fn = promiseClient[name];
    if (method.methodKind === "unary" || method.methodKind === "client_streaming") {
      result[name] = (request: unknown, options?: CallOptions) =>
        Effect.tryPromise({
          try: (signal) =>
            (fn as (req: unknown, opts: CallOptions) => Promise<unknown>)(request, {
              ...options,
              signal,
            }),
          catch: (err) => ConnectError.from(err),
        });
    } else {
      result[name] = fn;
    }
  }
  return result as EffectClient<Desc>;
}
