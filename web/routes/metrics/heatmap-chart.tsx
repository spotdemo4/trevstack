import type { HeatmapResponse } from "$gen/cursor/v1/metrics_pb";
import { Chart, type ChartDrawContext, styles } from "$lib/chart";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { scaleLinear, scaleSqrt } from "d3-scale";
import { type Component, createMemo } from "solid-js";

import { formatNumber } from "./format";

type Datum = { column: number; row: number; columns: number; rows: number; count: number };

const margin = { top: 8, right: 16, bottom: 28, left: 40 };

/** The pad's shape, so the heatmap lines up with the pads positions are sent from. */
const PAD_ASPECT_RATIO = 16 / 9;

const range = (index: number, cells: number) =>
  `${(index / cells).toFixed(2)}-${((index + 1) / cells).toFixed(2)}`;

const draw = ({ g, width, height, data, tooltip }: ChartDrawContext<Datum>) => {
  const { columns, rows } = data[0]!;
  const x = scaleLinear().domain([0, 1]).range([0, width]);
  const y = scaleLinear().domain([0, 1]).range([0, height]);
  // Square-root opacity keeps rarely visited cells visible next to the busiest ones.
  const opacity = scaleSqrt()
    .domain([0, max(data, (d) => d.count) ?? 1])
    .range([0, 1]);

  g.append("rect")
    .attr("class", "fill-sunken")
    .attr("width", width)
    .attr("height", height)
    .attr("rx", 4);

  const cells = g
    .append("g")
    .selectAll("rect")
    .data(data)
    .join("rect")
    .attr("class", `${styles.Grid} fill-chart-1 hover:fill-chart-1-hover`)
    .attr("x", (d) => x(d.column / columns))
    .attr("y", (d) => y(d.row / rows))
    .attr("width", width / columns)
    .attr("height", height / rows)
    .attr("fill-opacity", (d) => (d.count === 0 ? 0 : Math.max(0.06, opacity(d.count))));
  tooltip.attach(
    cells,
    (d) =>
      `x ${range(d.column, d.columns)}, y ${range(d.row, d.rows)}\nPositions: ${formatNumber(d.count)}`,
  );

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(
      axisBottom(x)
        .ticks(Math.max(2, Math.floor(width / 80)))
        .tickSizeOuter(0),
    );

  g.append("g")
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(
      axisLeft(y)
        .ticks(Math.max(2, Math.floor(height / 50)))
        .tickSizeOuter(0),
    );
};

/** Positions per cell of a grid over the pad, darker where cursors spend more time. */
export const HeatmapChart: Component<{ heatmap?: HeatmapResponse; emptyHeight: number }> = (
  props,
) => {
  const data = createMemo(() => {
    const heatmap = props.heatmap;
    if (!heatmap || !heatmap.counts.some((count) => count > 0n)) return [];
    const { columns, rows } = heatmap;
    return heatmap.counts.map((count, index) => ({
      column: index % columns,
      row: Math.floor(index / columns),
      columns,
      rows,
      count: Number(count),
    }));
  });

  return (
    <Chart
      data={data()}
      height={props.emptyHeight}
      aspectRatio={data().length > 0 ? PAD_ASPECT_RATIO : undefined}
      margin={margin}
      draw={draw}
      emptyLabel="No data in range"
    />
  );
};
