import { useLinkState, type TypedPath } from "@solidjs/router";
import { dynamic, type JSX } from "@solidjs/web";
import { omit } from "solid-js";
import { twMerge } from "tailwind-merge";

type NavLinkAsRouterProps = Omit<JSX.AnchorHTMLAttributes<HTMLAnchorElement>, "href" | "state"> & {
  as?: never;
  href: string | TypedPath;
  state?: unknown;
  activeClass?: string;
  inactiveClass?: string;
  end?: boolean;
};

type NavLinkAsAnchorProps = JSX.AnchorHTMLAttributes<HTMLAnchorElement> & {
  as: "a";
  activeClass?: never;
  end?: never;
  inactiveClass?: never;
};

type NavLinkProps = NavLinkAsRouterProps | NavLinkAsAnchorProps;

const baseClass =
  "inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text";
const activeClass = "text-ctp-text";
const Anchor = dynamic(() => "a");

export const NavLink = (props: NavLinkProps) => {
  if (props.as === "a") {
    const rest = omit(props, "as", "class");
    return <Anchor {...rest} class={twMerge(baseClass, props.class as never)} />;
  }

  const link = useLinkState(() => props.href, { end: props.end });
  const rest = omit(props, "as", "href", "state", "class", "activeClass", "inactiveClass", "end");

  return (
    <a
      {...rest}
      href={props.href}
      state={props.state === undefined ? undefined : JSON.stringify(props.state)}
      aria-current={link.current() ? "page" : undefined}
      data-active={link.active() ? "" : undefined}
      class={twMerge(
        baseClass,
        props.class as never,
        link.active()
          ? twMerge(activeClass, props.activeClass)
          : (props.inactiveClass ?? "inactive"),
      )}
    />
  );
};
