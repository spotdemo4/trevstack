import { SendRequestSchema } from "$gen/cursor/v1/send_pb";
import { Card } from "$lib/card";
import { CursorPad, type PadPoint, padPosition } from "$lib/cursor-pad";
import { errorDescription } from "$lib/effect";
import { Form } from "$lib/form/form";
import { useForm } from "$lib/form/hook";
import { NumberField } from "$lib/form/number-field";
import { protoSchema } from "$lib/form/proto-schema";
import { SubmitButton } from "$lib/form/submit-button";
import { PageHeader } from "$lib/page-header";
import { CursorClient } from "$lib/rpc";
import { toaster } from "$lib/toast";
import { Effect } from "effect";
import { type Component, Show } from "solid-js";

/** The fields' step, which the pad rounds to so its positions are valid inputs. */
const STEP = 0.01;

const inPad = (value: number | undefined): value is number =>
  value !== undefined && value >= 0 && value <= 1;

const Forms: Component = () => {
  const form = useForm(protoSchema(SendRequestSchema), () => ({
    validation: {
      onMount: true,
      onChange: true,
    },
    onSubmit: ({ value }) =>
      CursorClient.send(value).pipe(
        Effect.match({
          onSuccess: (response) => {
            toaster.success({
              title: "Position sent",
              description: `recorded as stream ${response.cursor?.streamId ?? "?"}`,
            });
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to send position",
              description: errorDescription(err),
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  const chosen = (): PadPoint | undefined => {
    const { x, y } = form.values();
    return inPad(x) && inPad(y) ? { x, y } : undefined;
  };

  const choose = (point: PadPoint) => {
    form.field("x").handleChange(Number(point.x.toFixed(2)));
    form.field("y").handleChange(Number(point.y.toFixed(2)));
  };

  return (
    <div class="mx-auto flex max-w-5xl flex-col gap-6 p-4">
      <PageHeader title="Forms" kind="Unary" rpc="rpc Send(SendRequest) returns (SendResponse)">
        Pick a point on the pad or type its coordinates, then send it. Coordinates are fractions of
        the pad's width and height from its top-left corner. Watch it arrive on the server streams
        page.
      </PageHeader>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-[1fr_16rem]">
        <CursorPad label="Position pad" onPress={choose} class="cursor-crosshair">
          <Show when={chosen()}>
            {(point) => (
              <div
                aria-hidden="true"
                class="pointer-events-none absolute size-4 -translate-1/2 rounded-full border-2 border-primary bg-primary/30"
                style={padPosition(point())}
              />
            )}
          </Show>
        </CursorPad>

        <Card class="p-4">
          <Form form={form}>
            <NumberField field={form.field("x")} label="X" min={0} max={1} step={STEP} />
            <NumberField field={form.field("y")} label="Y" min={0} max={1} step={STEP} />
            <SubmitButton form={form} label="Send" />
          </Form>
        </Card>
      </div>
    </div>
  );
};

export default Forms;
