import { Tabs } from "$lib/tabs";
import { render, type JSX } from "@solidjs/web";
import { createSignal, flush } from "solid-js";
import { afterEach, expect, test, vi } from "vitest";

import styles from "./tabs.module.css";

let dispose: (() => void) | undefined;

function mount(ui: () => JSX.Element) {
  const container = document.body.appendChild(document.createElement("div"));
  const disposeRender = render(ui, container);
  dispose = () => {
    disposeRender();
    container.remove();
  };
  return container;
}

function tab(container: HTMLElement, name: string) {
  const trigger = [...container.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
    (element) => element.textContent === name,
  );
  if (!trigger) throw new Error(`no tab named ${name}`);
  return trigger;
}

function visiblePanel(container: HTMLElement) {
  const panels = [...container.querySelectorAll<HTMLElement>('[role="tabpanel"]')];
  return panels.filter((panel) => !panel.hidden).map((panel) => panel.textContent);
}

function press(target: HTMLElement, key: string) {
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
  flush();
}

function Example(props: { value?: string; onValueChange?: (value: string) => void }) {
  return (
    <Tabs.Root defaultValue="one" value={props.value} onValueChange={props.onValueChange}>
      <Tabs.List aria-label="Example">
        <Tabs.Trigger value="one">One</Tabs.Trigger>
        <Tabs.Trigger value="two" disabled>
          Two
        </Tabs.Trigger>
        <Tabs.Trigger value="three">Three</Tabs.Trigger>
      </Tabs.List>
      <Tabs.Content value="one">First</Tabs.Content>
      <Tabs.Content value="two">Second</Tabs.Content>
      <Tabs.Content value="three">Third</Tabs.Content>
    </Tabs.Root>
  );
}

afterEach(() => {
  dispose?.();
  dispose = undefined;
});

test("links triggers to panels and shows only the selected one", () => {
  const container = mount(() => <Example />);
  const one = tab(container, "One");
  const panel = document.getElementById(one.getAttribute("aria-controls")!);

  expect(panel?.getAttribute("aria-labelledby")).toBe(one.id);
  expect(one.getAttribute("aria-selected")).toBe("true");
  expect(one.tabIndex).toBe(0);
  expect(tab(container, "Three").tabIndex).toBe(-1);
  expect(visiblePanel(container)).toEqual(["First"]);
});

test("clicking a trigger selects it", () => {
  const container = mount(() => <Example />);
  tab(container, "Three").click();
  flush();

  expect(tab(container, "Three").getAttribute("aria-selected")).toBe("true");
  expect(tab(container, "One").getAttribute("aria-selected")).toBe("false");
  expect(visiblePanel(container)).toEqual(["Third"]);
});

test("arrow keys skip disabled triggers and wrap around", () => {
  const container = mount(() => <Example />);
  const one = tab(container, "One");
  const three = tab(container, "Three");

  press(one, "ArrowRight");
  expect(document.activeElement).toBe(three);
  expect(visiblePanel(container)).toEqual(["Third"]);

  press(three, "ArrowRight");
  expect(document.activeElement).toBe(one);
  expect(visiblePanel(container)).toEqual(["First"]);

  press(one, "End");
  expect(visiblePanel(container)).toEqual(["Third"]);
  press(three, "Home");
  expect(visiblePanel(container)).toEqual(["First"]);
});

test("controlled tabs report changes without switching on their own", () => {
  const onValueChange = vi.fn();
  const [value, setValue] = createSignal("one");
  const container = mount(() => <Example value={value()} onValueChange={onValueChange} />);

  tab(container, "Three").click();
  flush();
  expect(onValueChange).toHaveBeenCalledExactlyOnceWith("three");
  expect(visiblePanel(container)).toEqual(["First"]);

  setValue("three");
  flush();
  expect(visiblePanel(container)).toEqual(["Third"]);

  tab(container, "Three").click();
  flush();
  expect(onValueChange).toHaveBeenCalledOnce();
});

test("indicator moves to the selected trigger", () => {
  const container = mount(() => <Example />);
  const indicator = container.querySelector<HTMLElement>("[data-tabs-indicator]")!;
  const three = tab(container, "Three");
  for (const [key, value] of Object.entries({
    offsetLeft: 120,
    offsetTop: 4,
    offsetWidth: 60,
    offsetHeight: 36,
  })) {
    Object.defineProperty(three, key, { configurable: true, value });
  }

  // Unmeasurable (zero-size) triggers leave the indicator hidden.
  expect(indicator.hidden).toBe(true);

  three.click();
  flush();
  expect(indicator.hidden).toBe(false);
  expect(indicator.style.transform).toBe("translate3d(120px, 4px, 0)");
  expect(indicator.style.width).toBe("60px");
  expect(indicator.style.height).toBe("36px");
});

test("panels only animate in after switching tabs", () => {
  const container = mount(() => <Example />);
  const panel = (name: string) =>
    document.getElementById(tab(container, name).getAttribute("aria-controls")!)!;

  expect(panel("One").classList.contains(styles.content)).toBe(false);

  tab(container, "Three").click();
  flush();
  expect(panel("Three").classList.contains(styles.content)).toBe(true);

  tab(container, "One").click();
  flush();
  expect(panel("One").classList.contains(styles.content)).toBe(true);
});
