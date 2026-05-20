import type { DistributionBucket } from "$connect/number/v1/metrics_pb";
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

type DistributionChartProps = {
  buckets: DistributionBucket[];
};

const margin = { top: 12, right: 16, bottom: 64, left: 44 };

export const DistributionChart: Component<DistributionChartProps> = (props) => {
  let containerRef!: HTMLDivElement;
  let svgRef!: SVGSVGElement;
  let tooltipRef!: HTMLDivElement;

  const { width, height } = useChartSize(() => containerRef, 260);
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
    svg.selectAll("*").remove();
    hideChartTooltip(tooltipRef);
    if (data.length === 0) return;

    const showTooltip = (event: PointerEvent, text: string) => {
      showChartTooltip({ event, container: containerRef, tooltip: tooltipRef, text });
    };
    const hideTooltip = () => {
      hideChartTooltip(tooltipRef);
    };

    const { innerW, innerH } = getChartInnerSize(w, h, margin);

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
    <ChartFrame
      width={width()}
      height={height()}
      isEmpty={props.buckets.length === 0}
      emptyLabel="No data in range"
      containerRef={setContainerRef}
      svgRef={setSvgRef}
      tooltipRef={setTooltipRef}
    />
  );
};
