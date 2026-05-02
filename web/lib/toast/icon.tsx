import { CircleAlert, CircleCheck, Info, LoaderCircle, TriangleAlert } from "lucide-solid";
import type { JSX } from "solid-js";

export const renderToastIcon = (type?: string): JSX.Element => {
  switch (type) {
    case "success":
      return <CircleCheck size={16} />;
    case "error":
      return <CircleAlert size={16} />;
    case "warning":
      return <TriangleAlert size={16} />;
    case "loading":
      return <LoaderCircle size={16} class="animate-spin" />;
    default:
      return <Info size={16} />;
  }
};
