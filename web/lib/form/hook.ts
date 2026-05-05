import { createFormHook } from "@tanstack/solid-form";

import { fieldContext, formContext } from "./context";
import { DateField } from "./DateField";
import { Form } from "./Form";
import { NumberField } from "./NumberField";
import { ResetButton } from "./ResetButton";
import { SubmitButton } from "./SubmitButton";
import { TextField } from "./TextField";

const { useAppForm, withForm } = createFormHook({
  fieldComponents: {
    TextField,
    NumberField,
    DateField,
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
