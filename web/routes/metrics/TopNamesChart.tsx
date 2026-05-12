import type { TopName } from "$connect/number/v1/metrics_pb";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { format } from "d3-format";
import { scaleBand, scaleLinear } from "d3-scale";
import { pointer, select } from "d3-selection";
import { type Component, createEffect, Show } from "solid-js";

import { useChartSize } from "./useChartSize";

type TopNamesChartProps = {
  names: TopName[];
};

const margin = { top: 8, right: 24, bottom: 28, left: 96 };
const rowHeight = 28;

const TopNamesChart: Component<TopNamesChartProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let containerRef!: HTMLDivElement;
  // oxlint-disable-next-line no-unassigned-vars
  let svgRef!: SVGSVGElement;
  // oxlint-disable-next-line no-unassigned-vars
  let tooltipRef!: HTMLDivElement;
  const dynamicHeight = () =>
    Math.max(120, props.names.length * rowHeight + margin.top + margin.bottom);
  const { width } = useChartSize(() => containerRef, 0);

  createEffect(() => {
    const w = width();
    const h = dynamicHeight();
    if (w === 0) return;

    const data = props.names.map((n) => ({
      name: n.name,
      count: Number(n.count),
      sum: Number(n.sum),
      average: n.average,
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
          `${d.name}\nTotal value: ${formatInteger(d.sum)}\nCount: ${formatInteger(d.count)}\nAverage: ${formatInteger(d.average)}`,
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

    const y = scaleBand<string>()
      .domain(data.map((d) => d.name))
      .range([0, innerH])
      .padding(0.2);

    const x = scaleLinear()
      .domain([0, max(data, (d) => d.count) ?? 1])
      .nice()
      .range([0, innerW]);

    g.append("g")
      .attr("class", "text-ctp-subtext0")
      .call(axisLeft(y).tickSizeOuter(0))
      .selectAll("text")
      .attr("class", "truncate");

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", "text-ctp-subtext0")
      .call(
        axisBottom(x)
          .ticks(Math.max(2, Math.floor(innerW / 80)))
          .tickSizeOuter(0),
      );

    g.append("g")
      .selectAll("rect")
      .data(data)
      .enter()
      .append("rect")
      .attr("class", "fill-ctp-peach")
      .attr("x", 0)
      .attr("y", (d) => y(d.name) ?? 0)
      .attr("width", (d) => x(d.count))
      .attr("height", y.bandwidth())
      .attr("rx", 2)
      .on("pointerenter", (event: PointerEvent, d) => {
        select(event.currentTarget as SVGRectElement).attr("class", "fill-ctp-yellow");
        showTooltip(event, d);
      })
      .on("pointermove", (event: PointerEvent, d) => {
        showTooltip(event, d);
      })
      .on("pointerleave", (event: PointerEvent) => {
        select(event.currentTarget as SVGRectElement).attr("class", "fill-ctp-peach");
        hideTooltip();
      });

    g.append("g")
      .selectAll("text")
      .data(data)
      .enter()
      .append("text")
      .attr("class", "pointer-events-none fill-ctp-text font-mono text-xs tabular-nums")
      .attr("x", (d) => x(d.count) + 6)
      .attr("y", (d) => (y(d.name) ?? 0) + y.bandwidth() / 2)
      .attr("dy", "0.35em")
      .text((d) => d.count);
  });

  return (
    <div ref={containerRef} class="relative w-full">
      <svg ref={svgRef} width={width()} height={dynamicHeight()} class="block" />
      <div
        ref={tooltipRef}
        class="pointer-events-none absolute z-10 max-w-56 rounded-md border border-ctp-surface1 bg-ctp-base/95 px-2 py-1 text-xs font-medium whitespace-pre text-ctp-text opacity-0 shadow-lg transition-opacity"
      />
      <Show when={props.names.length === 0}>
        <div class="absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0">
          No data in range
        </div>
      </Show>
    </div>
  );
};

export default TopNamesChart;
