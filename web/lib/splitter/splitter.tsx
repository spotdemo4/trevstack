import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { createContext, createSignal, omit, useContext } from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./splitter.module.css";

type PanelDefinition = {
  id: string;
  minSize?: number;
  maxSize?: number;
};

type SplitterContextValue = {
  root: () => HTMLDivElement | undefined;
  size: () => number;
  min: number;
  max: number;
  setSize: (size: number) => void;
};

const SplitterContext = createContext<SplitterContextValue>();

const useSplitter = () => {
  const context = useContext(SplitterContext);
  if (!context) throw new Error("Splitter components must be used within Splitter.Root");
  return context;
};

type RootProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> & {
  class?: string;
  defaultSize?: number[];
  panels: PanelDefinition[];
};

export const Root: Component<RootProps> = (props) => {
  const rest = omit(props, "defaultSize", "panels", "class", "children");
  const firstPanel = () => props.panels[0] ?? { id: "first" };
  const min = firstPanel().minSize ?? 10;
  const max = firstPanel().maxSize ?? 90;
  const clamp = (size: number) => Math.min(max, Math.max(min, size));
  const [size, setSize] = createSignal(clamp(props.defaultSize?.[0] ?? 25));
  let root: HTMLDivElement | undefined;

  return (
    <SplitterContext
      value={{ root: () => root, size, min, max, setSize: (next) => setSize(clamp(next)) }}
    >
      <div
        {...rest}
        ref={(node) => (root = node)}
        class={twMerge("grid min-w-0", props.class)}
        style={{ "grid-template-columns": `${size()}% 0.5rem minmax(0, 1fr)` }}
      >
        {props.children}
      </div>
    </SplitterContext>
  );
};

type PanelProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> & {
  class?: string;
  id: string;
};

export const Panel: Component<PanelProps> = (props) => {
  const rest = omit(props, "id", "class");
  return <div {...rest} id={props.id} class={twMerge("min-w-0", props.class)} />;
};

type ResizeTriggerProps = Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class"> & {
  class?: string;
  id: string;
};

export const ResizeTrigger: Component<ResizeTriggerProps> = (props) => {
  const splitter = useSplitter();
  const rest = omit(props, "id", "class");
  const updateFromPointer = (clientX: number) => {
    const root = splitter.root();
    if (!root) return;
    const rect = root.getBoundingClientRect();
    if (rect.width === 0) return;
    splitter.setSize(((clientX - rect.left) / rect.width) * 100);
  };

  return (
    <button
      {...rest}
      id={props.id}
      type="button"
      role="separator"
      aria-label={props["aria-label"] ?? "Resize panels"}
      aria-orientation="vertical"
      aria-valuemin={splitter.min}
      aria-valuemax={splitter.max}
      aria-valuenow={Math.round(splitter.size())}
      class={twMerge(styles.ResizeTrigger, props.class)}
      data-orientation="horizontal"
      onPointerDown={(event) => {
        event.currentTarget.setPointerCapture(event.pointerId);
        event.currentTarget.dataset.dragging = "";
        updateFromPointer(event.clientX);
      }}
      onPointerMove={(event) => {
        if (event.currentTarget.hasPointerCapture(event.pointerId)) {
          updateFromPointer(event.clientX);
        }
      }}
      onPointerUp={(event) => {
        event.currentTarget.releasePointerCapture(event.pointerId);
        delete event.currentTarget.dataset.dragging;
      }}
      onPointerCancel={(event) => {
        delete event.currentTarget.dataset.dragging;
      }}
      onKeyDown={(event) => {
        const step = event.shiftKey ? 10 : 1;
        if (event.key === "ArrowLeft") splitter.setSize(splitter.size() - step);
        else if (event.key === "ArrowRight") splitter.setSize(splitter.size() + step);
        else if (event.key === "Home") splitter.setSize(splitter.min);
        else if (event.key === "End") splitter.setSize(splitter.max);
        else return;
        event.preventDefault();
      }}
    >
      <span class={styles.ResizeTriggerIndicator} aria-hidden="true" />
    </button>
  );
};

export const Splitter = {
  Root,
  Panel,
  ResizeTrigger,
};
