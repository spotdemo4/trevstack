import { Skeleton } from "$lib/skeleton";
import { createFormHook } from "@tanstack/solid-form";
import { type Component, type ComponentProps, lazy, Show, Suspense } from "solid-js";
import { twMerge } from "tailwind-merge";

import { fieldContext, formContext } from "./context";
import type { DateField as DateFieldImpl } from "./DateField";
import { Form } from "./Form";
import type { NumberField as NumberFieldImpl } from "./NumberField";
import { ResetButton } from "./ResetButton";
import type { SelectField as SelectFieldImpl } from "./SelectField";
import { SubmitButton } from "./SubmitButton";
import type { TextField as TextFieldImpl } from "./TextField";

import styles from "./hook.module.css";

const FieldFallback: Component<{ label?: string; class?: string }> = (props) => (
  <div class={`${styles.fieldFallback} flex flex-col gap-1.5`}>
    <Show when={props.label}>
      <label class="text-sm font-medium text-ctp-subtext1">{props.label}</label>
    </Show>
    <Skeleton
      class={twMerge("h-9.5 min-w-42 border border-ctp-surface1 bg-ctp-base", props.class)}
    />
  </div>
);

const LazyTextField = lazy(() => import("./TextField").then((m) => ({ default: m.TextField })));

const TextField: Component<ComponentProps<typeof TextFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} />}>
    <div class={styles.fieldContent}>
      <LazyTextField {...props} />
    </div>
  </Suspense>
);

const LazyNumberField = lazy(() =>
  import("./NumberField").then((m) => ({ default: m.NumberField })),
);

const NumberField: Component<ComponentProps<typeof NumberFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} />}>
    <div class={styles.fieldContent}>
      <LazyNumberField {...props} />
    </div>
  </Suspense>
);

const LazyDateField = lazy(() => import("./DateField").then((m) => ({ default: m.DateField })));

const DateField: Component<ComponentProps<typeof DateFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} class={props.class} />}>
    <div class={styles.fieldContent}>
      <LazyDateField {...props} />
    </div>
  </Suspense>
);

const LazySelectField = lazy(() =>
  import("./SelectField").then((m) => ({ default: m.SelectField })),
);

const SelectField: Component<ComponentProps<typeof SelectFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} />}>
    <div class={styles.fieldContent}>
      <LazySelectField {...props} />
    </div>
  </Suspense>
);

const { useAppForm, withForm } = createFormHook({
  fieldComponents: {
    TextField,
    NumberField,
    DateField,
    SelectField,
  },
  formComponents: {
    Form,
    SubmitButton,
    ResetButton,
  },
  fieldContext,
  formContext,
});

export { useAppForm as useForm, withForm };
