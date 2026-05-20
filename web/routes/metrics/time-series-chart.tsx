import type { TimeSeriesPoint } from "$connect/number/v1/metrics_pb";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { extent, max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { format } from "d3-format";
import { scaleLinear, scaleTime } from "d3-scale";
import { pointer, select } from "d3-selection";
import { area, curveMonotoneX, line } from "d3-shape";
import { type Component, createEffect, Show } from "solid-js";

import { useChartSize } from "./use-chart-size";

import styles from "./chart-motion.module.css";

type TimeSeriesChartProps = {
  points: TimeSeriesPoint[];
};

const margin = { top: 12, right: 16, bottom: 28, left: 44 };

export const TimeSeriesChart: Component<TimeSeriesChartProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let containerRef!: HTMLDivElement;
  // oxlint-disable-next-line no-unassigned-vars
  let svgRef!: SVGSVGElement;
  // oxlint-disable-next-line no-unassigned-vars
  let tooltipRef!: HTMLDivElement;
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
        average: p.average,
      }));

    const svg = select(svgRef);
    const tooltip = select(tooltipRef);
    svg.selectAll("*").remove();
    tooltip.style("opacity", "0");
    if (data.length === 0) return;

    const formatInteger = format(",");
    const showTooltip = (event: PointerEvent, d: (typeof data)[number]) => {
      const [xPos, yPos] = pointer(event, containerRef);
      tooltip
        .text(
          `${d.date.toLocaleString()}\nTotal value: ${formatInteger(d.sum)}\nCount: ${formatInteger(d.count)}\nAverage: ${formatInteger(d.average)}`,
        )
        .style("opacity", "1");

      const tooltipNode = tooltipRef;
      const containerRect = containerRef.getBoundingClientRect();
      const tooltipRect = tooltipNode.getBoundingClientRect();
      const offset = 12;
      const left =
        xPos + tooltipRect.width + offset > containerRect.width
          ? xPos - tooltipRect.width - offset
          : xPos + offset;
      const top =
        yPos + tooltipRect.height + offset > containerRect.height
          ? yPos - tooltipRect.height - offset
          : yPos + offset;

      tooltip
        .style("left", `${Math.max(offset, left)}px`)
        .style("top", `${Math.max(offset, top)}px`);
    };
    const hideTooltip = () => {
      tooltip.style("opacity", "0");
    };

    const innerW = Math.max(0, w - margin.left - margin.right);
    const innerH = Math.max(0, h - margin.top - margin.bottom);

    const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

    const x = scaleTime()
      .domain(extent(data, (d) => d.date) as [Date, Date])
      .range([0, innerW]);

    const y = scaleLinear()
      .domain([0, max(data, (d) => d.count) ?? 1])
      .nice()
      .range([innerH, 0]);

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
      .call(
        axisBottom(x)
          .ticks(Math.max(2, Math.floor(innerW / 90)))
          .tickSizeOuter(0),
      );

    g.append("g")
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
      .call(axisLeft(y).ticks(5).tickSizeOuter(0));

    // Subtle horizontal grid lines.
    g.append("g")
      .attr("class", `${styles.Grid} text-ctp-surface1`)
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

    const areaGen = area<(typeof data)[number]>()
      .x((d) => x(d.date))
      .y0(innerH)
      .y1((d) => y(d.count))
      .curve(curveMonotoneX);

    const lineGen = line<(typeof data)[number]>()
      .x((d) => x(d.date))
      .y((d) => y(d.count))
      .curve(curveMonotoneX);

    g.append("path")
      .datum(data)
      .attr("class", `${styles.Area} fill-ctp-blue/20`)
      .attr("d", areaGen);

    g.append("path")
      .datum(data)
      .attr("class", `${styles.Line} stroke-ctp-blue`)
      .attr("fill", "none")
      .attr("pathLength", 1)
      .attr("stroke-width", 2)
      .attr("d", lineGen);

    g.append("g")
      .selectAll("circle")
      .data(data)
      .enter()
      .append("circle")
      .attr("class", `${styles.Point} fill-ctp-blue`)
      .attr("cx", (d) => x(d.date))
      .attr("cy", (d) => y(d.count))
      .attr("r", 3)
      .on("pointerenter", (event: PointerEvent, d) => {
        select(event.currentTarget as SVGCircleElement).attr("r", 5);
        showTooltip(event, d);
      })
      .on("pointermove", (event: PointerEvent, d) => {
        showTooltip(event, d);
      })
      .on("pointerleave", (event: PointerEvent) => {
        select(event.currentTarget as SVGCircleElement).attr("r", 3);
        hideTooltip();
      });
  });

  return (
    <div ref={containerRef} class="relative w-full">
      <svg ref={svgRef} width={width()} height={height()} class={`${styles.ChartCanvas} block`} />
      <div
        ref={tooltipRef}
        class="pointer-events-none absolute z-10 max-w-56 rounded-md border border-ctp-surface1 bg-ctp-base/95 px-2 py-1 text-xs font-medium whitespace-pre text-ctp-text opacity-0 shadow-lg transition-opacity"
      />
      <Show when={props.points.length === 0}>
        <div
          class={`${styles.EmptyState} absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0`}
        >
          No data in range
        </div>
      </Show>
    </div>
  );
};
