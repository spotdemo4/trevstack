import { CursorSchema, StreamEndSchema } from "$gen/cursor/v1/cursor_pb";
import { create } from "@bufbuild/protobuf";
import { flush } from "solid-js";
import { expect, test } from "vitest";

import { createCursorStore, type CursorEvent } from "./store.ts";

const moved = (streamId: bigint, x: number, y: number): CursorEvent => ({
  case: "moved",
  value: create(CursorSchema, { streamId, username: "alice", x, y }),
});

const ended = (streamId: bigint): CursorEvent => ({
  case: "ended",
  value: create(StreamEndSchema, { streamId, username: "alice" }),
});

function snapshot(store: ReturnType<typeof createCursorStore>) {
  flush();
  return store.cursors.map(({ id, username, x, y, trail }) => ({
    id,
    username,
    x,
    y,
    trail: trail.map((point) => ({ ...point })),
  }));
}

test("moves keep one cursor per stream with a bounded trail", () => {
  const store = createCursorStore({ trailLength: 2 });
  store.apply(moved(1n, 0.1, 0.2));
  store.apply(moved(2n, 0.5, 0.5));
  store.apply(moved(1n, 0.3, 0.4));
  store.apply(moved(1n, 0.6, 0.8));

  expect(snapshot(store)).toEqual([
    {
      id: "1",
      username: "alice",
      x: 0.6,
      y: 0.8,
      trail: [
        { x: 0.3, y: 0.4 },
        { x: 0.6, y: 0.8 },
      ],
    },
    { id: "2", username: "alice", x: 0.5, y: 0.5, trail: [{ x: 0.5, y: 0.5 }] },
  ]);
});

test("ended streams and idle cursors are removed", () => {
  let now = 0;
  const store = createCursorStore({ idleMs: 1000, now: () => now });
  store.apply(moved(1n, 0.1, 0.1));
  store.apply(moved(2n, 0.2, 0.2));
  store.apply(moved(3n, 0.3, 0.3));
  store.apply(ended(2n));
  store.apply({ case: undefined });

  now = 900;
  store.apply(moved(3n, 0.4, 0.4));
  now = 1500;
  store.prune();

  expect(snapshot(store).map((cursor) => cursor.id)).toEqual(["3"]);

  store.clear();
  expect(snapshot(store)).toEqual([]);
});
