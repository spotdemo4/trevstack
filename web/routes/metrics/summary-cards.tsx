import type { SummaryResponse } from "$gen/cursor/v1/metrics_pb";
import { Card } from "$lib/card";
import { Skeleton } from "$lib/skeleton";
import { type Component, For, Show } from "solid-js";

import { formatNumber } from "./format";

import styles from "./summary-cards.module.css";

type SummaryCardsProps = {
  /** Shows skeletons while undefined. */
  data?: SummaryResponse;
  /** Shows skeletons instead of stale data while refetching. */
  loading?: boolean;
};

const SUMMARY_CARD_COUNT = 6;

export const SummaryCards: Component<SummaryCardsProps> = (props) => {
  const stats = () => {
    const d = props.data;
    if (!d) return [];
    return [
      { label: "Positions", value: formatNumber(d.count) },
      { label: "Streams", value: formatNumber(d.streams) },
      { label: "Users", value: formatNumber(d.users) },
      { label: "Distance", value: formatNumber(d.distance) },
      { label: "Positions / stream", value: formatNumber(d.averageCount) },
      { label: "Distance / stream", value: formatNumber(d.averageDistance) },
    ];
  };
  const pending = () => props.loading || !props.data;

  return (
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
      <For each={Array.from({ length: SUMMARY_CARD_COUNT })}>
        {(_, index) => {
          const stat = () => stats()[index()];

          return (
            <div class={styles.Card} data-state={pending() ? "loading" : "ready"}>
              <Card class="p-4">
                <div class={styles.ContentFrame}>
                  <div class={styles.SkeletonLayer} aria-hidden={pending() ? "false" : "true"}>
                    <Skeleton class="mb-3 h-3 w-20" />
                    <Skeleton class="h-6 w-24" />
                  </div>

                  <Show when={stat()}>
                    {(s) => (
                      <div class={styles.ValueLayer} aria-hidden={pending() ? "true" : "false"}>
                        <div class="text-xs tracking-wide text-foreground-subtle uppercase">
                          {s().label}
                        </div>
                        <div class="mt-1 font-mono text-2xl text-foreground tabular-nums">
                          {s().value}
                        </div>
                      </div>
                    )}
                  </Show>
                </div>
              </Card>
            </div>
          );
        }}
      </For>
    </div>
  );
};
