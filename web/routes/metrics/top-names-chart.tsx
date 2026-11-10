import type { TopName } from "$gen/number/v1/metrics_pb";
import { Chart, type ChartDrawContext, styles } from "$lib/chart";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { scaleBand, scaleLinear } from "d3-scale";
import { type Component, createMemo } from "solid-js";

import { formatNumber, formatStats } from "./format";

type Datum = { name: string; count: number; sum: number; average: number };

const margin = { top: 8, right: 24, bottom: 28, left: 96 };
const rowHeight = 28;

const draw = ({ g, width, height, data, tooltip }: ChartDrawContext<Datum>) => {
  const y = scaleBand<string>()
    .domain(data.map((d) => d.name))
    .range([0, height])
    .padding(0.2);
  const x = scaleLinear()
    .domain([0, max(data, (d) => d.count) ?? 1])
    .nice()
    .range([0, width]);

  g.append("g")
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(axisLeft(y).tickSizeOuter(0))
    .selectAll("text")
    .attr("class", "truncate");

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(
      axisBottom(x)
        .ticks(Math.max(2, Math.floor(width / 80)))
        .tickSizeOuter(0),
    );

  const bars = g
    .append("g")
    .selectAll("rect")
    .data(data)
    .join("rect")
    .attr("class", `${styles.HorizontalBar} fill-chart-3 hover:fill-chart-3-hover`)
    .attr("x", 0)
    .attr("y", (d) => y(d.name) ?? 0)
    .attr("width", (d) => x(d.count))
    .attr("height", y.bandwidth())
    .attr("rx", 2);
  tooltip.attach(bars, (d) => formatStats(d.name, d));

  g.append("g")
    .selectAll("text")
    .data(data)
    .join("text")
    .attr(
      "class",
      `${styles.ValueLabel} pointer-events-none fill-foreground font-mono text-xs tabular-nums`,
    )
    .attr("x", (d) => x(d.count) + 6)
    .attr("y", (d) => (y(d.name) ?? 0) + y.bandwidth() / 2)
    .attr("dy", "0.35em")
    .text((d) => formatNumber(d.count));
};

/** Horizontal bar chart that grows past `minHeight` to fit one row per name. */
export const TopNamesChart: Component<{ names: TopName[]; minHeight: number }> = (props) => {
  const data = createMemo(() =>
    props.names.map((n) => ({
      name: n.name,
      count: Number(n.count),
      sum: Number(n.sum),
      average: n.average,
    })),
  );
  const height = () =>
    Math.max(props.minHeight, data().length * rowHeight + margin.top + margin.bottom);

  return (
    <Chart
      data={data()}
      height={height()}
      margin={margin}
      draw={draw}
      emptyLabel="No data in range"
    />
  );
};
