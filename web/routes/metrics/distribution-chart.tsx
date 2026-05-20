import type { DistributionBucket } from "$connect/number/v1/metrics_pb";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { format } from "d3-format";
import { scaleBand, scaleLinear } from "d3-scale";
import { pointer, select } from "d3-selection";
import { type Component, createEffect, Show } from "solid-js";

import { useChartSize } from "./use-chart-size";

import styles from "./chart-motion.module.css";

type DistributionChartProps = {
  buckets: DistributionBucket[];
};

const margin = { top: 12, right: 16, bottom: 64, left: 44 };

const DistributionChart: Component<DistributionChartProps> = (props) => {
  // oxlint-disable-next-line no-unassigned-vars
  let containerRef!: HTMLDivElement;
  // oxlint-disable-next-line no-unassigned-vars
  let svgRef!: SVGSVGElement;
  // oxlint-disable-next-line no-unassigned-vars
  let tooltipRef!: HTMLDivElement;
  const { width, height } = useChartSize(() => containerRef, 260);

  createEffect(() => {
    const w = width();
    const h = height();
    if (w === 0) return;

    const compactNumber = format("~s");
    const roundToNearestThousand = (value: number): number =>
      Math.abs(value) < 1000 ? value : Math.round(value / 1000) * 1000;

    const shortRangeLabel = (lower: number, upper: number): string => {
      const shortLower = compactNumber(roundToNearestThousand(lower)).replace("G", "B");
      const shortUpper = compactNumber(roundToNearestThousand(upper)).replace("G", "B");
      return lower === upper ? shortLower : `${shortLower}-${shortUpper}`;
    };

    const data = props.buckets.map((b, i) => ({
      key: `${i}`,
      label: shortRangeLabel(b.lower, b.upper),
      fullLabel: b.lower === b.upper ? `${b.lower}` : `${b.lower}-${b.upper}`,
      lower: b.lower,
      upper: b.upper,
      count: Number(b.count),
    }));

    const svg = select(svgRef);
    const tooltip = select(tooltipRef);
    svg.selectAll("*").remove();
    tooltip.style("opacity", "0");
    if (data.length === 0) return;

    const showTooltip = (event: PointerEvent, text: string) => {
      const [xPos, yPos] = pointer(event, containerRef);
      tooltip.text(text).style("opacity", "1");

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

    const x = scaleBand<string>()
      .domain(data.map((d) => d.key))
      .range([0, innerW])
      .padding(0.15);
    const labelByKey = new Map(data.map((d) => [d.key, d.label]));

    const y = scaleLinear()
      .domain([0, max(data, (d) => d.count) ?? 1])
      .nice()
      .range([innerH, 0]);

    // Show every Nth label so they don't collide on narrow widths.
    const skip = Math.max(1, Math.ceil(data.length / Math.max(1, Math.floor(innerW / 60))));

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
      .call(
        axisBottom(x)
          .tickValues(data.filter((_, i) => i % skip === 0).map((d) => d.key))
          .tickFormat((key) => labelByKey.get(key) ?? key)
          .tickSizeOuter(0),
      )
      .selectAll("text")
      .attr("transform", "rotate(-35)")
      .attr("dy", "0.35em")
      .style("text-anchor", "end");

    g.append("g")
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
      .call(axisLeft(y).ticks(5).tickSizeOuter(0));

    g.append("g")
      .selectAll("rect")
      .data(data)
      .enter()
      .append("rect")
      .attr("class", `${styles.VerticalBar} fill-ctp-mauve`)
      .attr("x", (d) => x(d.key) ?? 0)
      .attr("y", (d) => y(d.count))
      .attr("width", x.bandwidth())
      .attr("height", (d) => innerH - y(d.count))
      .attr("rx", 2)
      .on("pointerenter", (event: PointerEvent, d) => {
        select(event.currentTarget as SVGRectElement)
          .classed("fill-ctp-mauve", false)
          .classed("fill-ctp-pink", true);
        showTooltip(event, `${d.fullLabel}\nTotal count: ${d.count}`);
      })
      .on("pointermove", (event: PointerEvent, d) => {
        showTooltip(event, `${d.fullLabel}\nTotal count: ${d.count}`);
      })
      .on("pointerleave", (event: PointerEvent) => {
        select(event.currentTarget as SVGRectElement)
          .classed("fill-ctp-pink", false)
          .classed("fill-ctp-mauve", true);
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
      <Show when={props.buckets.length === 0}>
        <div
          class={`${styles.EmptyState} absolute inset-0 flex items-center justify-center text-sm text-ctp-subtext0`}
        >
          No data in range
        </div>
      </Show>
    </div>
  );
};

export default DistributionChart;
