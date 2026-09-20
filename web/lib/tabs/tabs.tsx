import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { createContext, createSignal, createUniqueId, omit, Show, useContext } from "solid-js";
import { twMerge } from "tailwind-merge";

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

export const List: Component<ListProps> = (props) => {
  const rest = omit(props, "class", "onKeyDown");

  return (
    <div
      {...rest}
      role="tablist"
      aria-orientation="horizontal"
      class={twMerge("inline-flex w-fit rounded-lg bg-ctp-crust p-1", props.class)}
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
    />
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
        "inline-flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium whitespace-nowrap text-ctp-subtext0 hover:text-ctp-text focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ctp-sky disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-ctp-surface0 data-[state=active]:font-semibold data-[state=active]:text-ctp-text data-[state=active]:shadow-sm",
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
        "min-w-0 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ctp-sky [[hidden]]:hidden",
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
