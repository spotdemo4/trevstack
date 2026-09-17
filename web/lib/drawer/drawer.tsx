import { X } from "$lib/icon";
import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { createContext, createUniqueId, omit, onCleanup, useContext } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./drawer.module.css";

const closeDuration = 250;
const openDuration = 320;

type ScrollLock = {
  bodyLeft: string;
  bodyOverflow: string;
  bodyPosition: string;
  bodyTop: string;
  bodyWidth: string;
  rootOverflow: string;
  scrollX: number;
  scrollY: number;
};

type DrawerContextValue = {
  titleId: string;
  open: () => void;
  close: () => void;
  handleClosed: () => void;
  setDialog: (dialog: HTMLDialogElement) => void;
  setTrigger: (trigger: HTMLButtonElement) => void;
};

const DrawerContext = createContext<DrawerContextValue>();

const useDrawer = () => {
  const context = useContext(DrawerContext);
  if (!context) throw new Error("Drawer components must be used within Drawer.Root");
  return context;
};

type RootProps = {
  children?: JSX.Element;
};

export const Root: Component<RootProps> = (props) => {
  const titleId = createUniqueId();
  let dialog: HTMLDialogElement | undefined;
  let trigger: HTMLButtonElement | undefined;
  let closeTimer: number | undefined;
  let openingFrame: number | undefined;
  let openingTimer: number | undefined;
  let scrollLock: ScrollLock | undefined;

  const cancelOpening = () => {
    if (openingFrame !== undefined) window.cancelAnimationFrame(openingFrame);
    if (openingTimer !== undefined) window.clearTimeout(openingTimer);
    openingFrame = undefined;
    openingTimer = undefined;
    dialog?.removeAttribute("data-opening");
    dialog?.removeAttribute("data-entering");
  };

  const lockScroll = () => {
    if (scrollLock) return;

    const root = document.documentElement;
    const body = document.body;
    scrollLock = {
      bodyLeft: body.style.left,
      bodyOverflow: body.style.overflow,
      bodyPosition: body.style.position,
      bodyTop: body.style.top,
      bodyWidth: body.style.width,
      rootOverflow: root.style.overflow,
      scrollX: window.scrollX,
      scrollY: window.scrollY,
    };

    root.style.overflow = "hidden";
    body.style.left = `-${scrollLock.scrollX}px`;
    body.style.overflow = "hidden";
    body.style.position = "fixed";
    body.style.top = `-${scrollLock.scrollY}px`;
    body.style.width = "100%";
  };

  const unlockScroll = () => {
    if (!scrollLock) return;

    const lock = scrollLock;
    scrollLock = undefined;
    const root = document.documentElement;
    const body = document.body;
    root.style.overflow = lock.rootOverflow;
    body.style.left = lock.bodyLeft;
    body.style.overflow = lock.bodyOverflow;
    body.style.position = lock.bodyPosition;
    body.style.top = lock.bodyTop;
    body.style.width = lock.bodyWidth;
    window.scrollTo(lock.scrollX, lock.scrollY);
  };

  const finishClose = () => {
    if (closeTimer !== undefined) window.clearTimeout(closeTimer);
    closeTimer = undefined;
    dialog?.removeAttribute("data-closing");
    if (dialog?.open) dialog.close();
  };

  const close = () => {
    if (!dialog?.open || dialog.hasAttribute("data-closing")) return;
    cancelOpening();
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      finishClose();
      return;
    }

    dialog.setAttribute("data-closing", "");
    closeTimer = window.setTimeout(finishClose, closeDuration);
  };

  const handleClosed = () => {
    if (closeTimer !== undefined) window.clearTimeout(closeTimer);
    closeTimer = undefined;
    cancelOpening();
    dialog?.removeAttribute("data-closing");
    unlockScroll();
    trigger?.focus();
  };

  onCleanup(() => {
    if (closeTimer !== undefined) window.clearTimeout(closeTimer);
    cancelOpening();
    unlockScroll();
  });

  return (
    <DrawerContext
      value={{
        titleId,
        open: () => {
          if (!dialog || dialog.open) return;
          dialog.removeAttribute("data-closing");
          const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
          lockScroll();
          try {
            dialog.showModal();
            if (!reducedMotion) {
              dialog.setAttribute("data-opening", "");
              dialog.scrollTop = 0;
            }
          } catch (error) {
            cancelOpening();
            unlockScroll();
            throw error;
          }
          if (!reducedMotion) {
            openingFrame = window.requestAnimationFrame(() => {
              openingFrame = window.requestAnimationFrame(() => {
                openingFrame = undefined;
                dialog?.removeAttribute("data-opening");
                dialog?.setAttribute("data-entering", "");
                openingTimer = window.setTimeout(() => {
                  openingTimer = undefined;
                  dialog?.removeAttribute("data-entering");
                }, openDuration);
              });
            });
          }
        },
        close,
        handleClosed,
        setDialog: (node) => (dialog = node),
        setTrigger: (node) => (trigger = node),
      }}
    >
      {props.children}
    </DrawerContext>
  );
};

type TriggerProps = Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class"> & {
  class?: string;
};

export const Trigger: Component<TriggerProps> = (props) => {
  const drawer = useDrawer();
  const rest = omit(props, "class", "children", "onClick");

  return (
    <button
      {...rest}
      ref={drawer.setTrigger}
      type={props.type ?? "button"}
      class={twMerge(
        "m-2 inline-flex cursor-pointer items-center gap-2 self-start rounded-md bg-ctp-surface0 px-3 py-1.5 text-sm font-medium text-ctp-text hover:bg-ctp-surface1",
        props.class,
      )}
      onClick={(event) => {
        if (typeof props.onClick === "function") props.onClick(event);
        drawer.open();
      }}
    >
      {props.children}
    </button>
  );
};

type ContentProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> & {
  backdropClass?: string;
  class?: string;
  positionerClass?: string;
};

export const Content: Component<ContentProps> = (props) => {
  const drawer = useDrawer();
  const contentProps = omit(props, "backdropClass", "positionerClass", "class");
  let content: HTMLDivElement | undefined;
  let dragStartY = 0;
  let dragDistance = 0;
  let lastDragY = 0;
  let lastDragTime = 0;
  let dragVelocity = 0;

  const setContent = (node: HTMLDivElement) => {
    content = node;
  };

  const setDragOffset = (offset: number) => {
    content?.style.setProperty("--drawer-drag-offset", `${offset}px`);
  };

  const settle = () => {
    if (!content) return;
    content.removeAttribute("data-dragging");
    content.setAttribute("data-settling", "");
    setDragOffset(0);
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      content.removeAttribute("data-settling");
      return;
    }
    content.addEventListener("transitionend", () => content?.removeAttribute("data-settling"), {
      once: true,
    });
  };

  const dismiss = () => {
    content?.removeAttribute("data-dragging");
    drawer.close();
  };

  return (
    <dialog
      ref={drawer.setDialog}
      aria-labelledby={drawer.titleId}
      class={twMerge(styles.dialog, props.backdropClass, props.positionerClass)}
      onClose={() => {
        content?.removeAttribute("data-dragging");
        content?.removeAttribute("data-settling");
        setDragOffset(0);
        drawer.handleClosed();
      }}
      onCancel={(event) => {
        event.preventDefault();
        drawer.close();
      }}
      onClick={(event) => {
        if (event.target === event.currentTarget) drawer.close();
      }}
    >
      <div
        {...contentProps}
        ref={setContent}
        class={twMerge(
          `${styles.content} flex max-h-[calc(100dvh-3rem)] w-full flex-col overflow-hidden rounded-t-xl border-t border-ctp-surface0 bg-ctp-mantle shadow-xl`,
          props.class,
        )}
      >
        <div
          aria-hidden="true"
          class="flex h-7 shrink-0 cursor-grab touch-none items-center justify-center active:cursor-grabbing"
          onPointerDown={(event) => {
            if (event.button !== 0) return;
            event.currentTarget.setPointerCapture(event.pointerId);
            dragStartY = event.clientY;
            dragDistance = 0;
            lastDragY = event.clientY;
            lastDragTime = event.timeStamp;
            dragVelocity = 0;
            content?.removeAttribute("data-settling");
            content?.setAttribute("data-dragging", "");
          }}
          onPointerMove={(event) => {
            if (!event.currentTarget.hasPointerCapture(event.pointerId)) return;
            const now = event.timeStamp;
            const elapsed = now - lastDragTime;
            dragDistance = Math.max(0, event.clientY - dragStartY);
            if (elapsed > 0) dragVelocity = (event.clientY - lastDragY) / elapsed;
            lastDragY = event.clientY;
            lastDragTime = now;
            setDragOffset(dragDistance);
          }}
          onPointerUp={(event) => {
            if (!event.currentTarget.hasPointerCapture(event.pointerId)) return;
            event.currentTarget.releasePointerCapture(event.pointerId);
            const distanceThreshold = Math.min((content?.offsetHeight ?? 0) * 0.25, 120);
            if (dragDistance > distanceThreshold || (dragDistance > 20 && dragVelocity > 0.5)) {
              dismiss();
            } else {
              settle();
            }
          }}
          onPointerCancel={settle}
        >
          <div class="h-1.5 w-12 rounded-full bg-ctp-surface2" />
        </div>
        <div class="flex min-h-0 flex-col gap-4 overflow-y-auto overscroll-contain px-4 pb-[calc(1rem+env(safe-area-inset-bottom))]">
          {props.children}
        </div>
      </div>
    </dialog>
  );
};

type TitleProps = Omit<JSX.HTMLAttributes<HTMLHeadingElement>, "class"> & {
  class?: string;
};

export const Title: Component<TitleProps> = (props) => {
  const drawer = useDrawer();
  return (
    <h2 {...props} id={drawer.titleId} class={twMerge("font-medium", props.class)}>
      {props.children}
    </h2>
  );
};

type CloseTriggerProps = Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class"> & {
  class?: string;
};

export const CloseTrigger: Component<CloseTriggerProps> = (props) => {
  const drawer = useDrawer();
  const rest = omit(props, "class", "children", "onClick");

  return (
    <button
      {...rest}
      type={props.type ?? "button"}
      class={twMerge(
        "cursor-pointer rounded-md p-1 text-ctp-subtext0 hover:bg-ctp-surface0 hover:text-ctp-text",
        props.class,
      )}
      onClick={(event) => {
        if (typeof props.onClick === "function") props.onClick(event);
        drawer.close();
      }}
    >
      {props.children ?? <X size={16} />}
    </button>
  );
};

export const Drawer = {
  Root,
  Trigger,
  Content,
  Title,
  CloseTrigger,
};
