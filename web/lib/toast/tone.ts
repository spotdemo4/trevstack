export const resolveToastTone = (type?: string) => {
  switch (type) {
    case "success":
      return {
        root: "border-ctp-green/35 bg-ctp-mantle",
        icon: "text-ctp-green",
        title: "text-ctp-text",
        progress: "bg-ctp-green/70",
      };
    case "error":
      return {
        root: "border-ctp-red/35 bg-ctp-mantle",
        icon: "border-ctp-red/35 bg-ctp-red/12 text-ctp-red",
        title: "text-ctp-text",
        progress: "bg-ctp-red/70",
      };
    case "warning":
      return {
        root: "border-ctp-yellow/35 bg-ctp-mantle",
        icon: "border-ctp-yellow/35 bg-ctp-yellow/12 text-ctp-yellow",
        title: "text-ctp-text",
        progress: "bg-ctp-yellow/70",
      };
    case "loading":
      return {
        root: "border-ctp-blue/35 bg-ctp-mantle",
        icon: "border-ctp-blue/35 bg-ctp-blue/12 text-ctp-blue",
        title: "text-ctp-text",
        progress: "bg-ctp-blue/70",
      };
    default:
      return {
        root: "border-ctp-sky/35 bg-ctp-mantle",
        icon: "border-ctp-sky/35 bg-ctp-sky/12 text-ctp-sky",
        title: "text-ctp-text",
        progress: "bg-ctp-sky/70",
      };
  }
};
