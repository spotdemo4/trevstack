import { Skeleton } from "$lib/skeleton";
import { createFormHook } from "@tanstack/solid-form";
import { type Component, type ComponentProps, lazy, Show, Suspense } from "solid-js";

import { fieldContext, formContext } from "./context";
import type { DateField as DateFieldImpl } from "./DateField";
import { Form } from "./Form";
import { NumberField } from "./NumberField";
import { ResetButton } from "./ResetButton";
import type { SelectField as SelectFieldImpl } from "./SelectField";
import { SubmitButton } from "./SubmitButton";
import { TextField } from "./TextField";

const FieldFallback: Component<{ label?: string }> = (props) => (
  <div class="flex flex-col gap-1.5">
    <Show when={props.label}>
      <Skeleton class="h-4 w-20" />
    </Show>
    <Skeleton class="h-9 min-w-42" />
  </div>
);

const LazyDateField = lazy(() => import("./DateField").then((m) => ({ default: m.DateField })));

const DateField: Component<ComponentProps<typeof DateFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} />}>
    <LazyDateField {...props} />
  </Suspense>
);

const LazySelectField = lazy(() =>
  import("./SelectField").then((m) => ({ default: m.SelectField })),
);

const SelectField: Component<ComponentProps<typeof SelectFieldImpl>> = (props) => (
  <Suspense fallback={<FieldFallback label={props.label} />}>
    <LazySelectField {...props} />
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
