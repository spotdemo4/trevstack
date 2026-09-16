import { X } from "$lib/icon";
import { Portal, type JSX } from "@solidjs/web";
import { createSignal, For, onCleanup, onSettled, Show, type Component } from "solid-js";

import { renderToastIcon } from "./icon";
import { toastController, toaster } from "./toaster";
import { resolveToastTone } from "./tone";

import styles from "./toast.module.css";

const STACK_GAP = 12;
const SWIPE_SLOP = 8;
const SWIPE_DIRECTION_RATIO = 1.25;
const SWIPE_MAX_DISTANCE = 120;
const SWIPE_DISTANCE_RATIO = 0.35;
const SWIPE_FLICK_DISTANCE = 24;
const SWIPE_FLICK_VELOCITY = 0.65;
const SWIPE_SAMPLE_WINDOW = 100;
const SWIPE_OFFSCREEN_CLEARANCE = 24;
const SWIPE_SNAPBACK_FALLBACK = 200;

type SwipePhase = "idle" | "tracking" | "dragging" | "snapback" | "committing";

type SwipeSample = {
  offset: number;
  time: number;
};

type SwipeSession = {
  pointerId: number;
  startX: number;
  startY: number;
  toastWidth: number;
  originalLeft: number;
  captured: boolean;
  samples: SwipeSample[];
};

const addSwipeSample = (session: SwipeSession, offset: number, time: number) => {
  session.samples.push({ offset, time });
  session.samples = session.samples.filter((sample) => sample.time >= time - SWIPE_SAMPLE_WINDOW);
};

const resolveSwipeVelocity = (samples: SwipeSample[]) => {
  const first = samples[0];
  const last = samples.at(-1);
  if (!first || !last || last.time <= first.time) return 0;
  return (last.offset - first.offset) / (last.time - first.time);
};

export const AppToaster: Component = () => {
  const [heights, setHeights] = createHeightMap();
  const observedIds = new WeakMap<Element, number>();
  const observer =
    typeof ResizeObserver === "undefined"
      ? undefined
      : new ResizeObserver((entries) => {
          for (const entry of entries) {
            const id = observedIds.get(entry.target);
            if (id !== undefined) updateHeight(id, (entry.target as HTMLElement).offsetHeight);
          }
        });

  let viewport: HTMLDivElement | undefined;
  let pointerInside = false;
  let focusInside = false;
  let activeSwipeId: number | undefined;

  const updateHeight = (id: number, height: number) => {
    setHeights((current) => {
      if (current.get(id) === height) return current;
      const next = new Map(current);
      next.set(id, height);
      return next;
    });
  };

  const observeToast = (id: number, element: HTMLDivElement) => {
    observedIds.set(element, id);
    observer?.observe(element);
    queueMicrotask(() => {
      if (observedIds.get(element) === id) updateHeight(id, element.offsetHeight);
    });
  };

  const unobserveToast = (id: number, element: HTMLDivElement) => {
    observer?.unobserve(element);
    observedIds.delete(element);
    queueMicrotask(() => {
      setHeights((current) => {
        if (!current.has(id)) return current;
        const next = new Map(current);
        next.delete(id);
        return next;
      });
    });
  };

  const resumeIfIdle = () => {
    if (!pointerInside && !focusInside && activeSwipeId === undefined) {
      toastController.resumeAll();
    }
  };

  const beginSwipe = (id: number) => {
    if (activeSwipeId !== undefined) return false;
    activeSwipeId = id;
    toastController.pauseAll();
    return true;
  };

  const endSwipe = (id: number) => {
    if (activeSwipeId !== id) return;
    activeSwipeId = undefined;
    resumeIfIdle();
  };

  onCleanup(() => {
    observer?.disconnect();
    activeSwipeId = undefined;
    queueMicrotask(() => toastController.resumeAll());
  });

  return (
    <Portal>
      <div
        ref={(element) => (viewport = element)}
        class={`${styles.toastViewport} pointer-events-none fixed z-100 w-[min(92vw,24rem)]`}
        onPointerEnter={(event) => {
          if (event.pointerType === "touch") return;
          pointerInside = true;
          toastController.pauseAll();
        }}
        onPointerLeave={(event) => {
          if (event.pointerType === "touch") return;
          pointerInside = false;
          resumeIfIdle();
        }}
        onFocusIn={() => {
          focusInside = true;
          toastController.pauseAll();
        }}
        onFocusOut={(event) => {
          if (event.relatedTarget instanceof Node && viewport?.contains(event.relatedTarget))
            return;
          focusInside = false;
          resumeIfIdle();
        }}
      >
        <For each={toaster.toasts()}>
          {(toast, index) => {
            const tone = () => resolveToastTone(toast.type);
            const stackIndex = () => toaster.toasts().length - 1 - index();
            const expandedOffset = () => {
              const newerToasts = toaster.toasts().slice(index() + 1);
              return newerToasts.reduce(
                (offset, newerToast) => offset + (heights().get(newerToast.id) ?? 0) + STACK_GAP,
                0,
              );
            };
            const [swipeX, setSwipeX] = createSignal(0);
            const [swipePhase, setSwipePhase] = createSignal<SwipePhase>("idle");
            let currentSwipePhase: SwipePhase = "idle";
            let toastElement: HTMLDivElement | undefined;
            let swipeElement: HTMLDivElement | undefined;
            let swipeSession: SwipeSession | undefined;
            let suppressNextClick = false;
            let clickSuppressionTimer: ReturnType<typeof setTimeout> | undefined;
            let snapbackTimer: ReturnType<typeof setTimeout> | undefined;

            const updateSwipePhase = (next: SwipePhase) => {
              currentSwipePhase = next;
              setSwipePhase(next);
            };

            const clearClickSuppressionTimer = () => {
              if (clickSuppressionTimer) clearTimeout(clickSuppressionTimer);
              clickSuppressionTimer = undefined;
            };

            const clearSnapbackTimer = () => {
              if (snapbackTimer) clearTimeout(snapbackTimer);
              snapbackTimer = undefined;
            };

            const clearClickSuppressionSoon = () => {
              clearClickSuppressionTimer();
              clickSuppressionTimer = setTimeout(() => {
                suppressNextClick = false;
                clickSuppressionTimer = undefined;
              });
            };

            const releasePointer = (session: SwipeSession) => {
              if (session.captured && swipeElement?.hasPointerCapture(session.pointerId)) {
                swipeElement.releasePointerCapture(session.pointerId);
              }
            };

            const scheduleIdleAfterSnapback = () => {
              clearSnapbackTimer();
              snapbackTimer = setTimeout(() => {
                if (currentSwipePhase === "snapback") updateSwipePhase("idle");
                snapbackTimer = undefined;
              }, SWIPE_SNAPBACK_FALLBACK);
            };

            const rejectTracking = () => {
              const session = swipeSession;
              swipeSession = undefined;
              if (session) releasePointer(session);
              setSwipeX(0);
              updateSwipePhase("idle");
              endSwipe(toast.id);
            };

            const cancelSwipe = () => {
              const session = swipeSession;
              if (!session) return;
              const wasDragging = currentSwipePhase === "dragging";
              swipeSession = undefined;
              releasePointer(session);

              if (wasDragging) {
                updateSwipePhase("snapback");
                setSwipeX(0);
                scheduleIdleAfterSnapback();
                clearClickSuppressionSoon();
              } else {
                setSwipeX(0);
                updateSwipePhase("idle");
              }

              endSwipe(toast.id);
            };

            const suppressDragClick = (event: MouseEvent) => {
              if (!suppressNextClick) return;
              event.preventDefault();
              event.stopPropagation();
              suppressNextClick = false;
              clearClickSuppressionTimer();
            };

            const finishSwipe = (event: PointerEvent) => {
              const session = swipeSession;
              if (!session || session.pointerId !== event.pointerId) return;

              if (currentSwipePhase !== "dragging") {
                swipeSession = undefined;
                setSwipeX(0);
                updateSwipePhase("idle");
                endSwipe(toast.id);
                return;
              }

              const offset = Math.max(0, event.clientX - session.startX);
              addSwipeSample(session, offset, event.timeStamp);
              const velocity = resolveSwipeVelocity(session.samples);
              const distanceThreshold = Math.min(
                SWIPE_MAX_DISTANCE,
                session.toastWidth * SWIPE_DISTANCE_RATIO,
              );
              const shouldDismiss =
                offset >= distanceThreshold ||
                (offset >= SWIPE_FLICK_DISTANCE && velocity >= SWIPE_FLICK_VELOCITY);

              swipeSession = undefined;
              releasePointer(session);
              clearClickSuppressionSoon();

              if (shouldDismiss) {
                updateSwipePhase("committing");
                setSwipeX(window.innerWidth - session.originalLeft + SWIPE_OFFSCREEN_CLEARANCE);
                toaster.dismiss(toast.id);
              } else {
                updateSwipePhase("snapback");
                setSwipeX(0);
                scheduleIdleAfterSnapback();
              }

              endSwipe(toast.id);
            };

            onSettled(() => {
              if (!toastElement || !swipeElement) return;
              observeToast(toast.id, toastElement);
              swipeElement.addEventListener("click", suppressDragClick, true);
            });

            onCleanup(() => {
              if (toastElement) unobserveToast(toast.id, toastElement);
              swipeElement?.removeEventListener("click", suppressDragClick, true);
              clearClickSuppressionTimer();
              clearSnapbackTimer();
              swipeSession = undefined;
              suppressNextClick = false;
              if (activeSwipeId === toast.id) {
                activeSwipeId = undefined;
                queueMicrotask(resumeIfIdle);
              }
            });

            return (
              <div
                data-frontmost={stackIndex() === 0 ? "true" : undefined}
                data-collapsed-hidden={stackIndex() >= 3 ? "true" : undefined}
                class={`${styles.toastItem} absolute inset-x-0 bottom-0`}
                style={
                  {
                    "--toast-index": stackIndex(),
                    "--toast-offset-y": `${expandedOffset()}px`,
                  } as JSX.CSSProperties
                }
              >
                <div
                  ref={(element) => (swipeElement = element)}
                  data-swipe-state={swipePhase() === "idle" ? undefined : swipePhase()}
                  class={styles.toastSwipe}
                  style={{ "--toast-swipe-x": `${swipeX()}px` } as JSX.CSSProperties}
                  onPointerDown={(event) => {
                    if (
                      event.pointerType !== "touch" ||
                      !event.isPrimary ||
                      event.button !== 0 ||
                      !toast.closable ||
                      stackIndex() !== 0 ||
                      toastController.state(toast.id) !== "open" ||
                      !beginSwipe(toast.id)
                    ) {
                      return;
                    }

                    clearClickSuppressionTimer();
                    clearSnapbackTimer();
                    suppressNextClick = false;
                    const rect = event.currentTarget.getBoundingClientRect();
                    const session: SwipeSession = {
                      pointerId: event.pointerId,
                      startX: event.clientX,
                      startY: event.clientY,
                      toastWidth: rect.width,
                      originalLeft: rect.left,
                      captured: false,
                      samples: [{ offset: 0, time: event.timeStamp }],
                    };
                    swipeSession = session;
                    event.currentTarget.setPointerCapture(event.pointerId);
                    session.captured = event.currentTarget.hasPointerCapture(event.pointerId);
                    setSwipeX(0);
                    updateSwipePhase("tracking");
                  }}
                  onPointerMove={(event) => {
                    const session = swipeSession;
                    if (!session || session.pointerId !== event.pointerId) return;

                    if (
                      !toast.closable ||
                      stackIndex() !== 0 ||
                      toastController.state(toast.id) !== "open"
                    ) {
                      cancelSwipe();
                      return;
                    }

                    const dx = event.clientX - session.startX;
                    const dy = event.clientY - session.startY;
                    const absoluteX = Math.abs(dx);
                    const absoluteY = Math.abs(dy);

                    if (currentSwipePhase === "tracking") {
                      if (Math.hypot(dx, dy) < SWIPE_SLOP) return;
                      if (dx <= 0) {
                        rejectTracking();
                        return;
                      }
                      if (absoluteX < absoluteY * SWIPE_DIRECTION_RATIO) {
                        if (absoluteY >= absoluteX) rejectTracking();
                        return;
                      }

                      suppressNextClick = true;
                      updateSwipePhase("dragging");
                    }

                    if (currentSwipePhase === "dragging") {
                      const offset = Math.max(0, dx);
                      setSwipeX(offset);
                      addSwipeSample(session, offset, event.timeStamp);
                    }
                  }}
                  onPointerUp={finishSwipe}
                  onPointerCancel={(event) => {
                    if (swipeSession?.pointerId === event.pointerId) cancelSwipe();
                  }}
                  onLostPointerCapture={(event) => {
                    if (
                      event.target === event.currentTarget &&
                      swipeSession?.pointerId === event.pointerId
                    ) {
                      cancelSwipe();
                    }
                  }}
                  onTransitionEnd={(event) => {
                    if (
                      event.target === event.currentTarget &&
                      event.propertyName === "transform" &&
                      currentSwipePhase === "snapback"
                    ) {
                      clearSnapbackTimer();
                      updateSwipePhase("idle");
                    }
                  }}
                >
                  <div
                    ref={(element) => (toastElement = element)}
                    role={toast.type === "error" ? "alert" : "status"}
                    aria-atomic="true"
                    data-state={toastController.state(toast.id)}
                    class={`${styles.toastRoot} relative w-full overflow-hidden rounded-lg border p-3 text-ctp-text shadow-lg shadow-ctp-crust/35 ${tone().root}`}
                    onAnimationEnd={(event) => {
                      if (
                        event.target === event.currentTarget &&
                        toastController.state(toast.id) === "closed"
                      ) {
                        toastController.finalize(toast.id);
                      }
                    }}
                  >
                    <div class={`${styles.toastContent} flex w-full items-start gap-3`}>
                      <div
                        class={`${styles.toastIcon} flex h-6 w-6 shrink-0 items-center justify-center ${tone().icon}`}
                      >
                        {renderToastIcon(toast.type)}
                      </div>

                      <div class="min-w-0 flex-1">
                        <div class={`truncate text-sm leading-5 font-semibold ${tone().title}`}>
                          {toast.title ?? "Notification"}
                        </div>
                        <Show when={toast.description}>
                          <div class="mt-1 text-sm leading-5 text-ctp-subtext1">
                            {toast.description}
                          </div>
                        </Show>
                      </div>

                      <Show when={toast.closable}>
                        <button
                          type="button"
                          aria-label="Dismiss notification"
                          class="mt-0.5 ml-1 inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-ctp-subtext0 transition-colors hover:bg-ctp-surface0/75 hover:text-ctp-text focus-visible:ring-2 focus-visible:ring-ctp-sky/40 focus-visible:outline-none"
                          onClick={() => toaster.dismiss(toast.id)}
                        >
                          <X size={14} />
                        </button>
                      </Show>
                    </div>

                    <div
                      class={`pointer-events-none absolute inset-x-0 bottom-0 h-0.5 ${tone().progress}`}
                    />
                  </div>
                </div>
              </div>
            );
          }}
        </For>
      </div>
    </Portal>
  );
};

const createHeightMap = () => {
  const [heights, setHeights] = createSignal(new Map<number, number>());
  return [heights, setHeights] as const;
};
