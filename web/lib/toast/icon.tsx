import CircleAlert from "lucide-solid/icons/circle-alert";
import CircleCheck from "lucide-solid/icons/circle-check";
import Info from "lucide-solid/icons/info";
import LoaderCircle from "lucide-solid/icons/loader-circle";
import TriangleAlert from "lucide-solid/icons/triangle-alert";
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
