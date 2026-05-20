import { pointer } from "d3-selection";

type ShowChartTooltipOptions = {
  event: PointerEvent;
  container: HTMLElement;
  tooltip: HTMLElement;
  text: string;
  offset?: number;
};

export function showChartTooltip(options: ShowChartTooltipOptions): void {
  const offset = options.offset ?? 12;
  const [xPos, yPos] = pointer(options.event, options.container);

  options.tooltip.textContent = options.text;
  options.tooltip.style.opacity = "1";

  const containerRect = options.container.getBoundingClientRect();
  const tooltipRect = options.tooltip.getBoundingClientRect();
  const left =
    xPos + tooltipRect.width + offset > containerRect.width
      ? xPos - tooltipRect.width - offset
      : xPos + offset;
  const top =
    yPos + tooltipRect.height + offset > containerRect.height
      ? yPos - tooltipRect.height - offset
      : yPos + offset;

  options.tooltip.style.left = `${Math.max(offset, left)}px`;
  options.tooltip.style.top = `${Math.max(offset, top)}px`;
}

export function hideChartTooltip(tooltip: HTMLElement): void {
  tooltip.style.opacity = "0";
}
