import type { JSX } from "@solidjs/web";
import { type Component, For, Show } from "solid-js";
import { twMerge } from "tailwind-merge";

import { userColor } from "./color";
import type { LiveCursor, PadPoint } from "./store";

import styles from "./pad.module.css";

type CursorPadProps = {
  /** Names the pad for assistive technology. */
  label: string;
  cursors?: LiveCursor[];
  class?: string;
  /** Called with every pointer move over the pad, including touch drags. */
  onMove?: (point: PadPoint) => void;
  /** Called when a pointer presses the pad. */
  onPress?: (point: PadPoint) => void;
  onLeave?: () => void;
  /** Drawn over the cursors, such as a marker positioned with `padPosition`. */
  children?: JSX.Element;
};

const clamp = (value: number) => Math.min(1, Math.max(0, value));

/** The position of a pointer event over `element`, clamped to the pad. */
export function padPoint(event: PointerEvent, element: Element): PadPoint {
  const rect = element.getBoundingClientRect();
  return {
    x: rect.width > 0 ? clamp((event.clientX - rect.left) / rect.width) : 0,
    y: rect.height > 0 ? clamp((event.clientY - rect.top) / rect.height) : 0,
  };
}

/** Styles that place an absolutely positioned element at `point` on the pad. */
export const padPosition = (point: PadPoint): JSX.CSSProperties => ({
  left: `${point.x * 100}%`,
  top: `${point.y * 100}%`,
});

const Trail: Component<{ cursor: LiveCursor }> = (props) => (
  <For each={props.cursor.trail.slice(1)} keyed={false}>
    {(point, index) => (
      <line
        x1={props.cursor.trail[index]?.x}
        y1={props.cursor.trail[index]?.y}
        x2={point().x}
        y2={point().y}
        stroke={userColor(props.cursor.username)}
        stroke-opacity={(index + 1) / props.cursor.trail.length}
        stroke-width="2"
        stroke-linecap="round"
        vector-effect="non-scaling-stroke"
      />
    )}
  </For>
);

const Marker: Component<{ cursor: LiveCursor }> = (props) => (
  <div
    class="pointer-events-none absolute"
    style={{ ...padPosition(props.cursor), color: userColor(props.cursor.username) }}
    data-cursor={props.cursor.username}
  >
    <svg
      aria-hidden="true"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      class="drop-shadow-[0_1px_1px_rgb(0_0_0/0.4)]"
    >
      <path
        d="M3 2l7.5 19 2.5-7.5L20.5 11z"
        fill="currentColor"
        stroke="var(--color-background)"
        stroke-width="1.5"
        stroke-linejoin="round"
      />
    </svg>
    <span class="absolute top-4 left-3 rounded-md bg-current px-1.5 py-0.5 text-xs font-medium whitespace-nowrap">
      <span class="text-background">{props.cursor.username}</span>
    </span>
  </div>
);

/**
 * A 16:9 area that shows live cursors with their trails, and reports the pointer's position on
 * it in fractions of its width and height, so pads of any size agree on positions.
 */
export const CursorPad: Component<CursorPadProps> = (props) => (
  <div
    role="group"
    aria-label={props.label}
    class={twMerge(
      styles.Pad,
      "relative aspect-video w-full touch-none overflow-hidden rounded-xl border border-line bg-sunken select-none",
      props.class,
    )}
    onPointerMove={(event) => props.onMove?.(padPoint(event, event.currentTarget))}
    onPointerDown={(event) => props.onPress?.(padPoint(event, event.currentTarget))}
    onPointerLeave={() => props.onLeave?.()}
  >
    <Show when={props.cursors}>
      {(cursors) => (
        <>
          <svg
            aria-hidden="true"
            class="pointer-events-none absolute inset-0 h-full w-full"
            viewBox="0 0 1 1"
            preserveAspectRatio="none"
          >
            <For each={cursors()}>{(cursor) => <Trail cursor={cursor} />}</For>
          </svg>
          <For each={cursors()}>{(cursor) => <Marker cursor={cursor} />}</For>
        </>
      )}
    </Show>
    {props.children}
  </div>
);
