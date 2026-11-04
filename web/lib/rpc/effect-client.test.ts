import {
  Code,
  type ResponseStream,
  type RpcServiceDescriptor,
  TrevRpcError,
} from "@trevrpc/trevrpc-js";
import { Effect, Stream } from "effect";
import { expect, test, vi } from "vitest";

import { createEffectClient } from "./effect-client.ts";

const service: RpcServiceDescriptor = {
  name: "TestService",
  fullName: "test.v1.TestService",
  exportName: "TestServiceService",
  methods: {
    get: { name: "Get", kind: "unary", inputType: "test.v1.Req", outputType: "test.v1.Res" },
    list: {
      name: "List",
      kind: "serverStreaming",
      inputType: "test.v1.Req",
      outputType: "test.v1.Res",
    },
  },
};

function responses(values: number[], error?: unknown): ResponseStream<number> {
  return {
    status: Promise.resolve({ code: Code.Ok, message: "", metadata: {} }),
    headers: Promise.resolve({}),
    close: async () => {},
    async *[Symbol.asyncIterator]() {
      yield* values;
      if (error) throw error;
    },
  };
}

test("unary calls pass the request and an abort signal", async () => {
  const get = vi.fn(async (request: { id: number }) => request.id * 2);
  const client = createEffectClient(service, { get, list: vi.fn() });

  await expect(Effect.runPromise(client.get({ id: 21 }, { timeoutMs: 5 }))).resolves.toBe(42);
  expect(get).toHaveBeenCalledWith(
    { id: 21 },
    expect.objectContaining({ timeoutMs: 5, signal: expect.any(AbortSignal) }),
  );
});

test("unary failures keep their TrevRpcError", async () => {
  const failure = new TrevRpcError(Code.NotFound, "missing");
  const client = createEffectClient(service, {
    get: async (): Promise<number> => {
      throw failure;
    },
    list: vi.fn(),
  });

  await expect(Effect.runPromise(Effect.flip(client.get({})))).resolves.toBe(failure);
});

test("other errors become Unknown TrevRpcErrors", async () => {
  const client = createEffectClient(service, {
    get: async (): Promise<number> => {
      throw new Error("boom");
    },
    list: vi.fn(),
  });

  const error = await Effect.runPromise(Effect.flip(client.get({})));
  expect(error).toBeInstanceOf(TrevRpcError);
  expect(error.code).toBe(Code.Unknown);
  expect(error.statusMessage).toBe("boom");
});

test("server streams yield every response", async () => {
  const client = createEffectClient(service, {
    get: vi.fn(),
    list: async () => responses([1, 2, 3]),
  });

  const items = await Effect.runPromise(Stream.runCollect(client.list({})));
  expect([...items]).toEqual([1, 2, 3]);
});

test("server stream failures keep their TrevRpcError", async () => {
  const failure = new TrevRpcError(Code.Unavailable, "connection lost");
  const client = createEffectClient(service, {
    get: vi.fn(),
    list: async () => responses([1], failure),
  });

  const error = await Effect.runPromise(Effect.flip(Stream.runCollect(client.list({}))));
  expect(error).toBe(failure);
});

test("interrupting a server stream aborts its call", async () => {
  let signal: AbortSignal | undefined;
  const client = createEffectClient(service, {
    get: vi.fn(),
    list: async (_request: unknown, options?: { signal?: AbortSignal }) => {
      signal = options?.signal;
      return responses([1, 2, 3]);
    },
  });

  await Effect.runPromise(Stream.runCollect(Stream.take(client.list({}), 1)));
  expect(signal?.aborted).toBe(true);
});
