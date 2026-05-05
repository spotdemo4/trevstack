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
    onSubmit: async ({ value }) => props.onSubmit(value),
  }));

  return (
    <form.AppForm>
      <form.Form class="justify-center">
        <form.AppField name="name">{(field) => <field.TextField label="Name" />}</form.AppField>
        <form.AppField name="min">{(field) => <field.NumberField label="Minimum" />}</form.AppField>
        <form.AppField name="max">{(field) => <field.NumberField label="Maximum" />}</form.AppField>
        <form.AppField name="start">{(field) => <field.DateField label="Start" />}</form.AppField>
        <form.AppField name="end">{(field) => <field.DateField label="End" />}</form.AppField>
        <form.SubmitButton />
      </form.Form>
    </form.AppForm>
  );
};

export default Form;
