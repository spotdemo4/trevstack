import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import {
  createContext,
  createEffect,
  createSignal,
  createUniqueId,
  omit,
  Show,
  useContext,
} from "solid-js";
import { twMerge } from "tailwind-merge";

import styles from "./tabs.module.css";

type TabsContextValue = {
  value: () => string | undefined;
  select: (value: string) => void;
  triggerId: (value: string) => string;
  contentId: (value: string) => string;
};

const TabsContext = createContext<TabsContextValue>();

const useTabs = () => {
  const context = useContext(TabsContext);
  if (!context) throw new Error("Tabs components must be used within Tabs.Root");
  return context;
};

type RootProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> & {
  class?: string;
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
};

export const Root: Component<RootProps> = (props) => {
  const rest = omit(props, "class", "children", "value", "defaultValue", "onValueChange");
  const id = createUniqueId();
  const [value, setValue] = createSignal(props.defaultValue);
  const selected = () => props.value ?? value();

  return (
    <TabsContext
      value={{
        value: selected,
        select: (next) => {
          if (next === selected()) return;
          if (props.value === undefined) setValue(next);
          props.onValueChange?.(next);
        },
        triggerId: (value) => `${id}-tab-${encodeURIComponent(value)}`,
        contentId: (value) => `${id}-panel-${encodeURIComponent(value)}`,
      }}
    >
      <div {...rest} class={twMerge("flex flex-col gap-6", props.class)}>
        {props.children}
      </div>
    </TabsContext>
  );
};

type ListProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> & {
  class?: string;
};

type IndicatorRect = { x: number; y: number; width: number; height: number };

export const List: Component<ListProps> = (props) => {
  const tabs = useTabs();
  const rest = omit(props, "class", "children", "onKeyDown");
  const [list, setList] = createSignal<HTMLDivElement>();
  const [indicator, setIndicator] = createSignal<IndicatorRect>();

  // Track the active trigger's box so the indicator can slide between triggers.
  createEffect(
    () => [list(), tabs.value()] as const,
    ([list, value]) => {
      const id = value === undefined ? undefined : tabs.triggerId(value);
      const trigger = list
        ? Array.from(list.querySelectorAll<HTMLElement>('[role="tab"]')).find((t) => t.id === id)
        : undefined;
      if (!list || !trigger) {
        setIndicator(undefined);
        return;
      }

      const measure = () =>
        setIndicator(
          trigger.offsetWidth > 0
            ? {
                x: trigger.offsetLeft,
                y: trigger.offsetTop,
                width: trigger.offsetWidth,
                height: trigger.offsetHeight,
              }
            : undefined,
        );
      measure();

      const observer = new ResizeObserver(measure);
      observer.observe(list);
      observer.observe(trigger);
      return () => observer.disconnect();
    },
  );

  return (
    <div
      {...rest}
      ref={setList}
      role="tablist"
      aria-orientation="horizontal"
      class={twMerge("relative inline-flex w-fit rounded-lg bg-sunken p-1", props.class)}
      onKeyDown={(event) => {
        if (typeof props.onKeyDown === "function") props.onKeyDown(event);
        if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey) return;
        if (!(event.target instanceof Element)) return;
        const list = event.currentTarget;
        const current = event.target.closest<HTMLButtonElement>('[role="tab"]');
        if (!current || current.closest('[role="tablist"]') !== list) return;

        const triggers = Array.from(
          list.querySelectorAll<HTMLButtonElement>('[role="tab"]:not(:disabled)'),
        ).filter((trigger) => trigger.closest('[role="tablist"]') === list);
        const index = triggers.indexOf(current);
        if (index === -1) return;

        const rtl = getComputedStyle(list).direction === "rtl";
        let next: number;
        if (event.key === "Home") next = 0;
        else if (event.key === "End") next = triggers.length - 1;
        else if (event.key === "ArrowRight") next = index + (rtl ? -1 : 1);
        else if (event.key === "ArrowLeft") next = index + (rtl ? 1 : -1);
        else return;

        event.preventDefault();
        const trigger = triggers[(next + triggers.length) % triggers.length];
        trigger?.focus();
        trigger?.click();
      }}
    >
      <div
        aria-hidden="true"
        data-tabs-indicator=""
        hidden={!indicator()}
        class={`${styles.indicator} pointer-events-none absolute top-0 left-0 rounded-md bg-fill-subtle shadow-sm`}
        style={{
          transform: `translate3d(${indicator()?.x ?? 0}px, ${indicator()?.y ?? 0}px, 0)`,
          width: `${indicator()?.width ?? 0}px`,
          height: `${indicator()?.height ?? 0}px`,
        }}
      />
      {props.children}
    </div>
  );
};

type TriggerProps = Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class" | "id" | "value"> & {
  class?: string;
  value: string;
};

export const Trigger: Component<TriggerProps> = (props) => {
  const tabs = useTabs();
  const rest = omit(props, "class", "value", "onClick");
  const active = () => tabs.value() === props.value;

  return (
    <button
      {...rest}
      type="button"
      id={tabs.triggerId(props.value)}
      role="tab"
      aria-selected={active() ? "true" : "false"}
      aria-controls={tabs.contentId(props.value)}
      tabindex={active() && !props.disabled ? 0 : -1}
      data-state={active() ? "active" : "inactive"}
      class={twMerge(
        "relative inline-flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium whitespace-nowrap text-foreground-subtle transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:pointer-events-none disabled:opacity-50 data-[state=active]:font-semibold data-[state=active]:text-foreground",
        props.class,
      )}
      onClick={(event) => {
        if (typeof props.onClick === "function") props.onClick(event);
        if (!event.defaultPrevented && !props.disabled) tabs.select(props.value);
      }}
    />
  );
};

type ContentProps = Omit<JSX.HTMLAttributes<HTMLDivElement>, "class" | "id"> & {
  class?: string;
  value: string;
};

export const Content: Component<ContentProps> = (props) => {
  const tabs = useTabs();
  const rest = omit(props, "class", "children", "value");
  const active = () => tabs.value() === props.value;
  // Only animate panels that have been hidden, so the initial panel doesn't fade in on load.
  const [animate, setAnimate] = createSignal(false);
  createEffect(active, (active) => {
    if (!active) setAnimate(true);
  });

  return (
    <div
      {...rest}
      id={tabs.contentId(props.value)}
      role="tabpanel"
      aria-labelledby={tabs.triggerId(props.value)}
      hidden={!active()}
      tabindex={0}
      data-state={active() ? "active" : "inactive"}
      class={twMerge(
        animate() && styles.content,
        "min-w-0 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring [[hidden]]:hidden",
        props.class,
      )}
    >
      <Show when={active()}>{props.children}</Show>
    </div>
  );
};

export const Tabs = {
  Root,
  List,
  Trigger,
  Content,
};
