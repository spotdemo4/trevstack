import type { JSX } from "@solidjs/web";
import type { Component } from "solid-js";
import { omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type IconProps = Omit<JSX.SvgSVGAttributes<SVGSVGElement>, "class"> & {
  class?: string;
  size?: number;
};

const Svg: Component<IconProps> = (props) => {
  const rest = omit(props, "size", "class", "children");
  const size = props.size ?? 24;

  return (
    <svg
      {...rest}
      aria-hidden={props["aria-label"] ? undefined : "true"}
      class={twMerge("shrink-0", props.class)}
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      {props.children}
    </svg>
  );
};

export const Moon: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" />
  </Svg>
);

export const Sun: Component<IconProps> = (props) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l1.41 1.41" />
  </Svg>
);

export const ExternalLink: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M15 3h6v6M10 14 21 3M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
  </Svg>
);

export const GitBranch: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M15 6a9 9 0 0 0-9 9V3" />
    <circle cx="18" cy="6" r="3" />
    <circle cx="6" cy="18" r="3" />
  </Svg>
);

export const User: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M20 21a8 8 0 0 0-16 0" />
    <circle cx="12" cy="7" r="4" />
  </Svg>
);

export const Ellipsis: Component<IconProps> = (props) => (
  <Svg {...props}>
    <circle cx="5" cy="12" r="1" />
    <circle cx="12" cy="12" r="1" />
    <circle cx="19" cy="12" r="1" />
  </Svg>
);

export const X: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M18 6 6 18M6 6l12 12" />
  </Svg>
);

export const Menu: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M4 6h16M4 12h16M4 18h16" />
  </Svg>
);

export const CircleAlert: Component<IconProps> = (props) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 8v4M12 16h.01" />
  </Svg>
);

export const CircleCheck: Component<IconProps> = (props) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="10" />
    <path d="m16 9-5.5 5.5L8 12" />
  </Svg>
);

export const Info: Component<IconProps> = (props) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 16v-4M12 8h.01" />
  </Svg>
);

export const LoaderCircle: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
  </Svg>
);

export const TriangleAlert: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
    <path d="M12 9v4M12 17h.01" />
  </Svg>
);

export const SlidersHorizontal: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M10 5H3M12 19H3M14 3v4M16 17v4M21 12h-9M21 19h-5M21 5h-7M8 10v4M8 12H3" />
  </Svg>
);

export const ChevronDown: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="m6 9 6 6 6-6" />
  </Svg>
);

export const ChevronLeft: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="m15 18-6-6 6-6" />
  </Svg>
);

export const ChevronRight: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="m9 18 6-6-6-6" />
  </Svg>
);

export const Calendar: Component<IconProps> = (props) => (
  <Svg {...props}>
    <path d="M8 2v4M16 2v4M3 10h18" />
    <rect width="18" height="18" x="3" y="4" rx="2" />
  </Svg>
);
