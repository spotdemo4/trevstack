import type { DistributionBucket } from "$gen/number/v1/metrics_pb";
import { Chart, type ChartDrawContext, styles } from "$lib/chart";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { scaleBand, scaleLinear } from "d3-scale";
import { type Component, createMemo } from "solid-js";

import { formatNumber, formatRange, formatRangeShort } from "./format";

type Datum = { key: string; label: string; fullLabel: string; count: number };

const margin = { top: 12, right: 16, bottom: 64, left: 44 };

const draw = ({ g, width, height, data, tooltip }: ChartDrawContext<Datum>) => {
  const x = scaleBand<string>()
    .domain(data.map((d) => d.key))
    .range([0, width])
    .padding(0.15);
  const y = scaleLinear()
    .domain([0, max(data, (d) => d.count) ?? 1])
    .nice()
    .range([height, 0]);

  // Label every Nth bar so labels don't collide on narrow widths.
  const labelEvery = Math.max(1, Math.ceil(data.length / Math.max(1, Math.floor(width / 60))));
  const labelByKey = new Map(data.map((d) => [d.key, d.label]));

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(
      axisBottom(x)
        .tickValues(data.filter((_, i) => i % labelEvery === 0).map((d) => d.key))
        .tickFormat((key) => labelByKey.get(key) ?? key)
        .tickSizeOuter(0),
    )
    .selectAll("text")
    .attr("transform", "rotate(-35)")
    .attr("dy", "0.35em")
    .style("text-anchor", "end");

  g.append("g")
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(axisLeft(y).ticks(5).tickSizeOuter(0));

  const bars = g
    .append("g")
    .selectAll("rect")
    .data(data)
    .join("rect")
    .attr("class", `${styles.VerticalBar} fill-chart-2 hover:fill-chart-2-hover`)
    .attr("x", (d) => x(d.key) ?? 0)
    .attr("y", (d) => y(d.count))
    .attr("width", x.bandwidth())
    .attr("height", (d) => height - y(d.count))
    .attr("rx", 2);
  tooltip.attach(bars, (d) => `${d.fullLabel}\nTotal count: ${formatNumber(d.count)}`);
};

export const DistributionChart: Component<{ buckets: DistributionBucket[]; height: number }> = (
  props,
) => {
  const data = createMemo(() =>
    props.buckets.map((b, i) => ({
      key: `${i}`,
      label: formatRangeShort(b.lower, b.upper),
      fullLabel: formatRange(b.lower, b.upper),
      count: Number(b.count),
    })),
  );

  return (
    <Chart
      data={data()}
      height={props.height}
      margin={margin}
      draw={draw}
      emptyLabel="No data in range"
    />
  );
};
