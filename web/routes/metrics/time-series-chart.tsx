import type { TimeSeriesPoint } from "$gen/cursor/v1/metrics_pb";
import { Chart, type ChartDrawContext, styles } from "$lib/chart";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { extent, max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { scaleLinear, scaleTime } from "d3-scale";
import { select } from "d3-selection";
import { area, curveMonotoneX, line } from "d3-shape";
import { type Component, createMemo } from "solid-js";

import { formatCounts } from "./format";

type Datum = { date: Date; count: number; streams: number; users: number };

const margin = { top: 12, right: 16, bottom: 28, left: 44 };

const draw = ({ g, width, height, data, tooltip }: ChartDrawContext<Datum>) => {
  const x = scaleTime()
    .domain(extent(data, (d) => d.date) as [Date, Date])
    .range([0, width]);
  const y = scaleLinear()
    .domain([0, max(data, (d) => d.count) ?? 1])
    .nice()
    .range([height, 0]);

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(
      axisBottom(x)
        .ticks(Math.max(2, Math.floor(width / 90)))
        .tickSizeOuter(0),
    );

  g.append("g")
    .attr("class", `${styles.Axis} text-foreground-subtle`)
    .call(axisLeft(y).ticks(5).tickSizeOuter(0));

  g.append("g")
    .attr("class", `${styles.Grid} text-line`)
    .selectAll("line")
    .data(y.ticks(5))
    .join("line")
    .attr("x1", 0)
    .attr("x2", width)
    .attr("y1", (d) => y(d))
    .attr("y2", (d) => y(d))
    .attr("stroke", "currentColor")
    .attr("stroke-dasharray", "2,3");

  g.append("path")
    .datum(data)
    .attr("class", `${styles.Area} fill-chart-1/20`)
    .attr(
      "d",
      area<Datum>()
        .x((d) => x(d.date))
        .y0(height)
        .y1((d) => y(d.count))
        .curve(curveMonotoneX),
    );

  g.append("path")
    .datum(data)
    .attr("class", `${styles.Line} stroke-chart-1`)
    .attr("fill", "none")
    .attr("pathLength", 1)
    .attr("stroke-width", 2)
    .attr(
      "d",
      line<Datum>()
        .x((d) => x(d.date))
        .y((d) => y(d.count))
        .curve(curveMonotoneX),
    );

  const points = g
    .append("g")
    .selectAll("circle")
    .data(data)
    .join("circle")
    .attr("class", `${styles.Point} fill-chart-1`)
    .attr("cx", (d) => x(d.date))
    .attr("cy", (d) => y(d.count))
    .attr("r", 3)
    .on("pointerenter", (event: PointerEvent) => {
      select(event.currentTarget as SVGCircleElement).attr("r", 5);
    })
    .on("pointerleave", (event: PointerEvent) => {
      select(event.currentTarget as SVGCircleElement).attr("r", 3);
    });
  tooltip.attach(points, (d) => formatCounts(d.date.toLocaleString(), d));
};

export const TimeSeriesChart: Component<{ points: TimeSeriesPoint[]; height: number }> = (
  props,
) => {
  const data = createMemo(() =>
    props.points.flatMap((p) =>
      p.bucket
        ? [
            {
              date: timestampDate(p.bucket),
              count: Number(p.count),
              streams: Number(p.streams),
              users: Number(p.users),
            },
          ]
        : [],
    ),
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
