import { Card } from "$lib/card";
import type { JSX } from "@solidjs/web";
import { Errored, Loading, type ParentComponent } from "solid-js";

type ChartCardProps = {
  title: string;
  /** Height of the loading and error placeholders, matching the chart's. */
  height: number;
  errorLabel: string;
  /** Rendered beside the title. */
  controls?: JSX.Element;
  class?: string;
};

/** Card with a title row and loading and error states for an async chart. */
export const ChartCard: ParentComponent<ChartCardProps> = (props) => {
  const Placeholder = (placeholder: { class: string; children: string }) => (
    <div
      class={`flex items-center justify-center text-sm ${placeholder.class}`}
      style={{ height: `${props.height}px` }}
    >
      {placeholder.children}
    </div>
  );

  return (
    <Card class={`p-4 ${props.class ?? ""}`}>
      <div class="mb-3 flex items-center justify-between gap-2">
        <h2 class="font-semibold text-foreground">{props.title}</h2>
        {props.controls}
      </div>
      <Errored fallback={<Placeholder class="text-danger">{props.errorLabel}</Placeholder>}>
        <Loading fallback={<Placeholder class="text-foreground-subtle">Loading...</Placeholder>}>
          {props.children}
        </Loading>
      </Errored>
    </Card>
  );
};
