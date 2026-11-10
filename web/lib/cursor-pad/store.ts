import type { Cursor } from "$gen/cursor/v1/cursor_pb";
import type { ShareResponse } from "$gen/cursor/v1/share_pb";
import type { WatchResponse } from "$gen/cursor/v1/watch_pb";
import { createStore, getOwner, onCleanup } from "solid-js";

/** A position on the pad, in fractions of its width and height from its top-left corner. */
export type PadPoint = { x: number; y: number };

export type LiveCursor = PadPoint & {
  /** The cursor's stream ID, or any other key unique among the store's cursors. */
  id: string;
  username: string;
  /** The latest positions, oldest first, ending with the current one. */
  trail: PadPoint[];
  /** When the cursor last moved, in milliseconds since the epoch. */
  updatedAt: number;
};

/** A cursor event of a Watch or Share response. */
export type CursorEvent = WatchResponse["event"] | ShareResponse["event"];

export type CursorStoreOptions = {
  /** How long a cursor may stay still before it is removed. 10 seconds by default. */
  idleMs?: number;
  /** How many positions each trail keeps, including the current one. 24 by default. */
  trailLength?: number;
  now?: () => number;
};

export type CursorStore = {
  /** The cursors, in the order they appeared. */
  cursors: LiveCursor[];
  /** Applies a cursor event of a Watch or Share response. */
  apply: (event: CursorEvent) => void;
  /** Moves the cursor `id`, adding it if it is new. */
  move: (id: string, username: string, point: PadPoint) => void;
  /** Removes the cursor `id`. */
  remove: (id: string) => void;
  /** Removes the cursors that have been still for longer than the idle time. */
  prune: () => void;
  clear: () => void;
};

const PRUNE_INTERVAL_MS = 1000;

/**
 * Keeps the live cursors of a pad by the events of Watch and Share streams. Created inside a
 * component, it also removes idle cursors every second until the component is cleaned up.
 */
export function createCursorStore(options: CursorStoreOptions = {}): CursorStore {
  const idleMs = options.idleMs ?? 10_000;
  const trailLength = Math.max(1, options.trailLength ?? 24);
  const now = options.now ?? Date.now;
  const [cursors, setCursors] = createStore<LiveCursor[]>([]);

  const move = (id: string, username: string, point: PadPoint) => {
    const updatedAt = now();
    setCursors((draft) => {
      const cursor = draft.find((candidate) => candidate.id === id);
      if (!cursor) {
        draft.push({ id, username, ...point, trail: [point], updatedAt });
        return;
      }
      cursor.x = point.x;
      cursor.y = point.y;
      cursor.updatedAt = updatedAt;
      cursor.trail.push(point);
      if (cursor.trail.length > trailLength)
        cursor.trail.splice(0, cursor.trail.length - trailLength);
    });
  };

  const remove = (id: string) => setCursors((draft) => draft.filter((cursor) => cursor.id !== id));

  const moveCursor = (cursor: Cursor) =>
    move(cursor.streamId.toString(), cursor.username, { x: cursor.x, y: cursor.y });

  const prune = () => {
    const cutoff = now() - idleMs;
    setCursors((draft) => draft.filter((cursor) => cursor.updatedAt >= cutoff));
  };

  if (getOwner()) {
    const interval = setInterval(prune, PRUNE_INTERVAL_MS);
    onCleanup(() => clearInterval(interval));
  }

  return {
    cursors,
    apply: (event) => {
      if (event.case === "moved") moveCursor(event.value);
      else if (event.case === "ended") remove(event.value.streamId.toString());
    },
    move,
    remove,
    prune,
    clear: () => setCursors(() => []),
  };
}
