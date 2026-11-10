import { Effect, Stream } from "effect";
import { expect, test } from "vitest";

import { createPushStream } from "./push-stream.ts";

test("streams the values pushed before and after it is consumed, until it ends", async () => {
  const { stream, push, end } = createPushStream<number>();
  push(1);
  push(2);

  const collected = Effect.runPromise(Stream.runCollect(stream));
  push(3);
  end();
  push(4);

  expect(await collected).toEqual([1, 2, 3]);
});

test("an ended stream that was never pushed to is empty", async () => {
  const { stream, end } = createPushStream<string>();
  end();

  expect(await Effect.runPromise(Stream.runCollect(stream))).toEqual([]);
});
