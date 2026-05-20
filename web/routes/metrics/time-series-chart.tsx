import type { TimeSeriesPoint } from "$connect/number/v1/metrics_pb";
import {
  ChartFrame,
  getChartInnerSize,
  hideChartTooltip,
  showChartTooltip,
  styles,
  useChartSize,
} from "$lib/chart";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { extent, max } from "d3-array";
import { axisBottom, axisLeft } from "d3-axis";
import { format } from "d3-format";
import { scaleLinear, scaleTime } from "d3-scale";
import { select } from "d3-selection";
import { area, curveMonotoneX, line } from "d3-shape";
import { type Component, createEffect } from "solid-js";

type TimeSeriesChartProps = {
  points: TimeSeriesPoint[];
};

const margin = { top: 12, right: 16, bottom: 28, left: 44 };

export const TimeSeriesChart: Component<TimeSeriesChartProps> = (props) => {
  let containerRef!: HTMLDivElement;
  let svgRef!: SVGSVGElement;
  let tooltipRef!: HTMLDivElement;

  const { width, height } = useChartSize(() => containerRef, 280);
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
    svg.selectAll("*").remove();
    hideChartTooltip(tooltipRef);
    if (data.length === 0) return;

    const formatInteger = format(",");
    const showTooltip = (event: PointerEvent, d: (typeof data)[number]) => {
      showChartTooltip({
        event,
        container: containerRef,
        tooltip: tooltipRef,
        text: `${d.date.toLocaleString()}\nTotal value: ${formatInteger(d.sum)}\nCount: ${formatInteger(d.count)}\nAverage: ${formatInteger(d.average)}`,
      });
    };
    const hideTooltip = () => {
      hideChartTooltip(tooltipRef);
    };

    const { innerW, innerH } = getChartInnerSize(w, h, margin);

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
    <ChartFrame
      width={width()}
      height={height()}
      isEmpty={props.points.length === 0}
      emptyLabel="No data in range"
      containerRef={setContainerRef}
      svgRef={setSvgRef}
      tooltipRef={setTooltipRef}
    />
  );
};
