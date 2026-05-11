import { Field, type FieldInputProps } from "@ark-ui/solid/field";
import { type Component, splitProps } from "solid-js";
import { twMerge } from "tailwind-merge";

type TextInputProps = FieldInputProps;

export const TextInput: Component<TextInputProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <Field.Input
      {...rest}
      class={twMerge(
        "rounded-md border border-ctp-surface1 bg-ctp-base px-3 py-2 text-sm text-ctp-text transition-colors placeholder:text-ctp-overlay0 hover:border-ctp-surface2 focus:border-ctp-sky focus:ring-2 focus:ring-ctp-sky/40 focus:outline-none data-invalid:border-ctp-red data-invalid:focus:ring-ctp-red/40",
        local.class,
      )}
    />
  );
};
