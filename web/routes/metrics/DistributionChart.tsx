import type { DistributionBucket } from "$connect/number/v1/metrics_pb";
import * as d3 from "d3";
import { type Component, createEffect, Show } from "solid-js";

import { useChartSize } from "./useChartSize";

type DistributionChartProps = {
  buckets: DistributionBucket[];
};

const margin = { top: 12, right: 16, bottom: 36, left: 44 };

const DistributionChart: Component<DistributionChartProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let containerRef!: HTMLDivElement;
  // oxlint-disable-next-line no-unassigned-vars
  let svgRef!: SVGSVGElement;
  const { width, height } = useChartSize(() => containerRef, 260);

  createEffect(() => {
    const w = width();
    const h = height();
    if (w === 0) return;

    const data = props.buckets.map((b) => ({
      label: b.lower === b.upper ? `${b.lower}` : `${b.lower}–${b.upper}`,
      lower: b.lower,
      upper: b.upper,
      count: Number(b.count),
    }));

    const svg = d3.select(svgRef);
    svg.selectAll("*").remove();
    if (data.length === 0) return;

    const innerW = Math.max(0, w - margin.left - margin.right);
    const innerH = Math.max(0, h - margin.top - margin.bottom);

    const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

    const x = d3
      .scaleBand<string>()
      .domain(data.map((d) => d.label))
      .range([0, innerW])
      .padding(0.15);

    const y = d3
      .scaleLinear()
      .domain([0, d3.max(data, (d) => d.count) ?? 1])
      .nice()
      .range([innerH, 0]);

    // Show every Nth label so they don't collide on narrow widths.
    const skip = Math.max(1, Math.ceil(data.length / Math.max(1, Math.floor(innerW / 60))));

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", "text-ctp-subtext0")
      .call(
        d3
          .axisBottom(x)
          .tickValues(data.filter((_, i) => i % skip === 0).map((d) => d.label))
          .tickSizeOuter(0),
      )
      .selectAll("text")
      .attr("transform", "rotate(-25)")
      .style("text-anchor", "end");

    g.append("g").attr("class", "text-ctp-subtext0").call(d3.axisLeft(y).ticks(5).tickSizeOuter(0));

    g.append("g")
      .selectAll("rect")
      .data(data)
      .enter()
      .append("rect")
      .attr("class", "fill-ctp-mauve")
      .attr("x", (d) => x(d.label) ?? 0)
      .attr("y", (d) => y(d.count))
      .attr("width", x.bandwidth())
      .attr("height", (d) => innerH - y(d.count))
      .attr("rx", 2)
      .append("title")
      .text((d) => `${d.label}: ${d.count}`);
  });

  return (
    <div ref={containerRef} class="relative w-full">
      <svg ref={svgRef} width={width()} height={height()} class="block" />
      <Show when={props.buckets.length === 0}>
        <div class="absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0">
          No data in range
        </div>
      </Show>
    </div>
  );
};

export default DistributionChart;
