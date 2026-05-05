import type { TimeSeriesPoint } from "$connect/number/v1/metrics_pb";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import * as d3 from "d3";
import { type Component, createEffect, Show } from "solid-js";

import { useChartSize } from "./useChartSize";

type TimeSeriesChartProps = {
  points: TimeSeriesPoint[];
};

const margin = { top: 12, right: 16, bottom: 28, left: 44 };

const TimeSeriesChart: Component<TimeSeriesChartProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let containerRef!: HTMLDivElement;
  // oxlint-disable-next-line no-unassigned-vars
  let svgRef!: SVGSVGElement;
  const { width, height } = useChartSize(() => containerRef, 280);

  createEffect(() => {
    const w = width();
    const h = height();
    if (w === 0) return;

    const data = props.points
      .filter((p) => p.bucket)
      .map((p) => ({
        date: timestampDate(p.bucket!),
        count: Number(p.count),
        sum: Number(p.sum),
      }));

    const svg = d3.select(svgRef);
    svg.selectAll("*").remove();
    if (data.length === 0) return;

    const innerW = Math.max(0, w - margin.left - margin.right);
    const innerH = Math.max(0, h - margin.top - margin.bottom);

    const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

    const x = d3
      .scaleTime()
      .domain(d3.extent(data, (d) => d.date) as [Date, Date])
      .range([0, innerW]);

    const y = d3
      .scaleLinear()
      .domain([0, d3.max(data, (d) => d.count) ?? 1])
      .nice()
      .range([innerH, 0]);

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", "text-ctp-subtext0")
      .call(
        d3
          .axisBottom(x)
          .ticks(Math.max(2, Math.floor(innerW / 90)))
          .tickSizeOuter(0),
      );

    g.append("g").attr("class", "text-ctp-subtext0").call(d3.axisLeft(y).ticks(5).tickSizeOuter(0));

    // Subtle horizontal grid lines.
    g.append("g")
      .attr("class", "text-ctp-surface1")
      .selectAll("line")
      .data(y.ticks(5))
      .enter()
      .append("line")
      .attr("x1", 0)
      .attr("x2", innerW)
      .attr("y1", (d) => y(d))
      .attr("y2", (d) => y(d))
      .attr("stroke", "currentColor")
      .attr("stroke-dasharray", "2,3");

    const area = d3
      .area<(typeof data)[number]>()
      .x((d) => x(d.date))
      .y0(innerH)
      .y1((d) => y(d.count))
      .curve(d3.curveMonotoneX);

    const line = d3
      .line<(typeof data)[number]>()
      .x((d) => x(d.date))
      .y((d) => y(d.count))
      .curve(d3.curveMonotoneX);

    g.append("path").datum(data).attr("class", "fill-ctp-blue/20").attr("d", area);

    g.append("path")
      .datum(data)
      .attr("class", "stroke-ctp-blue")
      .attr("fill", "none")
      .attr("stroke-width", 2)
      .attr("d", line);

    g.append("g")
      .selectAll("circle")
      .data(data)
      .enter()
      .append("circle")
      .attr("class", "fill-ctp-blue")
      .attr("cx", (d) => x(d.date))
      .attr("cy", (d) => y(d.count))
      .attr("r", 3)
      .append("title")
      .text((d) => `${d.date.toLocaleString()}\ncount: ${d.count}\nsum: ${d.sum}`);
  });

  return (
    <div ref={containerRef} class="relative w-full">
      <svg ref={svgRef} width={width()} height={height()} class="block" />
      <Show when={props.points.length === 0}>
        <div class="absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0">
          No data in range
        </div>
      </Show>
    </div>
  );
};

export default TimeSeriesChart;
