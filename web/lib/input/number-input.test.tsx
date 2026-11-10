import { render } from "@solidjs/web";
import { createSignal, flush } from "solid-js";
import { afterEach, expect, test } from "vitest";

import { NumberInput } from "./number-input";

let dispose: (() => void) | undefined;

afterEach(() => {
  dispose?.();
  dispose = undefined;
});

test("the value is written unless the input already holds that number", () => {
  const [value, setValue] = createSignal("0.5");
  const container = document.body.appendChild(document.createElement("div"));
  const disposeRender = render(() => <NumberInput value={value()} />, container);
  dispose = () => {
    disposeRender();
    container.remove();
  };
  flush();
  const input = container.querySelector("input")!;
  expect(input.value).toBe("0.5");

  setValue("0.25");
  flush();
  expect(input.value).toBe("0.25");

  // as typed, before the field reports the number
  input.value = "0.50";
  setValue("0.5");
  flush();
  expect(input.value).toBe("0.50");

  setValue("");
  flush();
  expect(input.value).toBe("");
});
