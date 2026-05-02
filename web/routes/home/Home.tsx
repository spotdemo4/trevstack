import type { AddRequest } from "$connect/number/v1/add_pb";
import { Card } from "$lib/card";
import { toaster } from "$lib/toast";
import { NumberClient } from "$lib/transport";
import type { Component } from "solid-js";

import Form from "./Form";

const Home: Component = () => {
  const onSubmit = async (value: AddRequest) => {
    const [response, err] = await NumberClient.add({
      name: value.name,
      number: value.number,
    });

    if (err) {
      toaster.error({
        title: "Failed to add number",
        description: err.message,
        closable: true,
      });
      return;
    }

    toaster.success({
      title: "Number added",
      description: `new sum: ${response.sum}`,
      closable: true,
    });
  };

  return (
    <div class="h-body flex flex-col items-center justify-center gap-4">
      <h1 class="text-2xl font-bold">Add Numbers</h1>
      <Card class="w-full max-w-sm">
        <Form onSubmit={onSubmit} />
      </Card>
    </div>
  );
};

export default Home;
