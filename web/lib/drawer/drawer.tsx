import { X } from "$lib/icon";
import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { createContext, createUniqueId, omit, useContext } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./drawer.module.css";

type DrawerContextValue = {
  titleId: string;
  open: () => void;
  close: () => void;
  restoreFocus: () => void;
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

  const close = () => dialog?.close();

  return (
    <DrawerContext
      value={{
        titleId,
        open: () => dialog?.showModal(),
        close,
        restoreFocus: () => trigger?.focus(),
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

  return (
    <dialog
      ref={drawer.setDialog}
      aria-labelledby={drawer.titleId}
      class={twMerge(styles.dialog, props.backdropClass, props.positionerClass)}
      onClose={drawer.restoreFocus}
      onClick={(event) => {
        if (event.target === event.currentTarget) drawer.close();
      }}
    >
      <div
        {...contentProps}
        class={twMerge(
          `${styles.content} flex h-full w-80 max-w-[85vw] flex-col gap-4 overflow-y-auto bg-ctp-mantle p-4 shadow-xl`,
          props.class,
        )}
      />
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
