import type { SummaryResponse } from "$connect/number/v1/metrics_pb";
import { Card } from "$lib/card";
import { Skeleton } from "$lib/skeleton";
import { type Component, For, Show } from "solid-js";

import styles from "./summary-cards.module.css";

type SummaryCardsProps = {
  data: SummaryResponse | undefined;
  loading: boolean;
};

const SUMMARY_CARD_COUNT = 6;

const formatBig = (n: bigint | number): string => Number(n).toLocaleString();
const formatFloat = (n: number): string =>
  n.toLocaleString(undefined, { maximumFractionDigits: 2 });

const SummaryCards: Component<SummaryCardsProps> = (props) => {
  const stats = () => {
    const d = props.data;
    if (!d) return [];
    return [
      { label: "Items", value: formatBig(d.totalCount) },
      { label: "Sum", value: formatBig(d.totalSum) },
      { label: "Average", value: formatFloat(d.average) },
      { label: "Min", value: formatBig(d.min) },
      { label: "Max", value: formatBig(d.max) },
      { label: "Distinct names", value: formatBig(d.distinctNames) },
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
                  <div class={styles.SkeletonLayer} aria-hidden={!pending()}>
                    <Skeleton class="mb-3 h-3 w-20" />
                    <Skeleton class="h-6 w-24" />
                  </div>

                  <Show when={stat()}>
                    {(s) => (
                      <div class={styles.ValueLayer} aria-hidden={pending()}>
                        <div class="text-xs tracking-wide text-ctp-subtext0 uppercase">
                          {s().label}
                        </div>
                        <div class="mt-1 font-mono text-2xl text-ctp-text tabular-nums">
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

export default SummaryCards;
