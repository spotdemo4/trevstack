import type { TopName } from "$connect/number/v1/metrics_pb";
import {
  ChartFrame,
  getChartInnerSize,
  hideChartTooltip,
  showChartTooltip,
  styles,
  useChartSize,
} from "$lib/chart";
import { max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { format } from "d3-format";
import { scaleBand, scaleLinear } from "d3-scale";
import { select } from "d3-selection";
import { type Component, createEffect } from "solid-js";

type TopNamesChartProps = {
  names: TopName[];
};

const margin = { top: 8, right: 24, bottom: 28, left: 96 };
const rowHeight = 28;
const minChartHeight = 260;

export const TopNamesChart: Component<TopNamesChartProps> = (props) => {
  let containerRef!: HTMLDivElement;
  let svgRef!: SVGSVGElement;
  let tooltipRef!: HTMLDivElement;

  const dynamicHeight = () =>
    Math.max(minChartHeight, props.names.length * rowHeight + margin.top + margin.bottom);
  const { width } = useChartSize(() => containerRef, 0);
  const setContainerRef = (element: HTMLDivElement) => {
    containerRef = element;
  };
  const setSvgRef = (element: SVGSVGElement) => {
    svgRef = element;
  };
  const setTooltipRef = (element: HTMLDivElement) => {
    tooltipRef = element;
  };

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
    svg.selectAll("*").remove();
    hideChartTooltip(tooltipRef);
    if (data.length === 0) return;

    const formatInteger = format(",");
    const showTooltip = (event: PointerEvent, d: (typeof data)[number]) => {
      showChartTooltip({
        event,
        container: containerRef,
        tooltip: tooltipRef,
        text: `${d.name}\nTotal value: ${formatInteger(d.sum)}\nCount: ${formatInteger(d.count)}\nAverage: ${formatInteger(d.average)}`,
      });
    };
    const hideTooltip = () => {
      hideChartTooltip(tooltipRef);
    };

    const { innerW, innerH } = getChartInnerSize(w, h, margin);

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
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
      .call(axisLeft(y).tickSizeOuter(0))
      .selectAll("text")
      .attr("class", "truncate");

    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .attr("class", `${styles.Axis} text-ctp-subtext0`)
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
      .attr("class", `${styles.HorizontalBar} fill-ctp-peach`)
      .attr("x", 0)
      .attr("y", (d) => y(d.name) ?? 0)
      .attr("width", (d) => x(d.count))
      .attr("height", y.bandwidth())
      .attr("rx", 2)
      .on("pointerenter", (event: PointerEvent, d) => {
        select(event.currentTarget as SVGRectElement)
          .classed("fill-ctp-peach", false)
          .classed("fill-ctp-yellow", true);
        showTooltip(event, d);
      })
      .on("pointermove", (event: PointerEvent, d) => {
        showTooltip(event, d);
      })
      .on("pointerleave", (event: PointerEvent) => {
        select(event.currentTarget as SVGRectElement)
          .classed("fill-ctp-yellow", false)
          .classed("fill-ctp-peach", true);
        hideTooltip();
      });

    g.append("g")
      .selectAll("text")
      .data(data)
      .enter()
      .append("text")
      .attr(
        "class",
        `${styles.ValueLabel} pointer-events-none fill-ctp-text font-mono text-xs tabular-nums`,
      )
      .attr("x", (d) => x(d.count) + 6)
      .attr("y", (d) => (y(d.name) ?? 0) + y.bandwidth() / 2)
      .attr("dy", "0.35em")
      .text((d) => d.count);
  });

  return (
    <ChartFrame
      width={width()}
      height={dynamicHeight()}
      isEmpty={props.names.length === 0}
      emptyLabel="No data in range"
      containerRef={setContainerRef}
      svgRef={setSvgRef}
      tooltipRef={setTooltipRef}
    />
  );
};
