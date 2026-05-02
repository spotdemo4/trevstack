import { type ListRequest, ListRequestSchema } from "$connect/number/v1/list_pb";
import { useForm } from "$lib/form/hook";
import { create } from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import type { Component } from "solid-js";

const Form: Component<{
  onSubmit: (value: ListRequest) => void;
}> = (props) => {
  const form = useForm(() => ({
    defaultValues: create(ListRequestSchema),
    validators: {
      onChange: createStandardSchema(ListRequestSchema),
    },
    onSubmit: async ({ value }) => {
      props.onSubmit(value);
    },
  }));

  return (
    <form.AppForm>
      <form.Form class="justify-center">
        <form.AppField name="name" children={(field) => <field.TextField label="Name" />} />
        <form.AppField name="min" children={(field) => <field.NumberField label="Minimum" />} />
        <form.AppField name="max" children={(field) => <field.NumberField label="Maximum" />} />
        <form.SubmitButton />
      </form.Form>
    </form.AppForm>
  );
};

export default Form;
