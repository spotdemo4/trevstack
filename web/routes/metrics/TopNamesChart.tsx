import type { TopName } from "$connect/number/v1/metrics_pb";
import * as d3 from "d3";
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
    }));

    const svg = d3.select(svgRef);
    svg.selectAll("*").remove();
    if (data.length === 0) return;

    const innerW = Math.max(0, w - margin.left - margin.right);
    const innerH = Math.max(0, h - margin.top - margin.bottom);

    const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

    const y = d3
      .scaleBand<string>()
      .domain(data.map((d) => d.name))
      .range([0, innerH])
      .padding(0.2);

    const x = d3
      .scaleLinear()
      .domain([0, d3.max(data, (d) => d.count) ?? 1])
      .nice()
      .range([0, innerW]);

    g.append("g")
      .attr("class", "text-ctp-subtext0")
      .call(d3.axisLeft(y).tickSizeOuter(0))
      .selectAll("text")
      .attr("class", "truncate");

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", "text-ctp-subtext0")
      .call(
        d3
          .axisBottom(x)
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
      .append("title")
      .text((d) => `${d.name}: count ${d.count}, sum ${d.sum}`);

    g.append("g")
      .selectAll("text")
      .data(data)
      .enter()
      .append("text")
      .attr("class", "fill-ctp-text font-mono text-xs tabular-nums")
      .attr("x", (d) => x(d.count) + 6)
      .attr("y", (d) => (y(d.name) ?? 0) + y.bandwidth() / 2)
      .attr("dy", "0.35em")
      .text((d) => d.count);
  });

  return (
    <div ref={containerRef} class="relative w-full">
      <svg ref={svgRef} width={width()} height={dynamicHeight()} class="block" />
      <Show when={props.names.length === 0}>
        <div class="absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0">
          No data in range
        </div>
      </Show>
    </div>
  );
};

export default TopNamesChart;
