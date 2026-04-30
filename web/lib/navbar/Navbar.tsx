import { useLocation } from "@solidjs/router";
import {
	type Accessor,
	type Component,
	createContext,
	createEffect,
	createSignal,
	type JSX,
	onCleanup,
	onMount,
	useContext,
} from "solid-js";
import { twMerge } from "tailwind-merge";

type NavbarProps = {
	children?: JSX.Element;
	class?: string;
};

type NavbarContextValue = {
	indicatorStyle: Accessor<JSX.CSSProperties>;
};

const NavbarContext = createContext<NavbarContextValue>();

const NavbarRoot: Component<NavbarProps> = (props) => {
	const location = useLocation();
	const [indicatorStyle, setIndicatorStyle] = createSignal<JSX.CSSProperties>({
		transform: "translateX(0px)",
		width: "0px",
		opacity: "0",
	});
	let navRef: HTMLElement | undefined;

	const updateIndicator = () => {
		if (!navRef) {
			return;
		}

		const activeLink = navRef.querySelector<HTMLAnchorElement>(
			'a[aria-current="page"], a.text-ctp-text',
		);
		if (!activeLink) {
			setIndicatorStyle((prevStyle) => ({
				...prevStyle,
				opacity: "0",
			}));
			return;
		}

		const navRect = navRef.getBoundingClientRect();
		const activeRect = activeLink.getBoundingClientRect();
		setIndicatorStyle({
			transform: `translateX(${activeRect.left - navRect.left}px)`,
			width: `${activeRect.width}px`,
			opacity: "1",
		});
	};

	onMount(() => {
		updateIndicator();

		if (!navRef) {
			return;
		}

		const resizeObserver = new ResizeObserver(updateIndicator);
		resizeObserver.observe(navRef);
		const navLinks = navRef.querySelectorAll("a");
		navLinks.forEach((link) => {
			resizeObserver.observe(link);
		});

		window.addEventListener("resize", updateIndicator);
		onCleanup(() => {
			resizeObserver.disconnect();
			window.removeEventListener("resize", updateIndicator);
		});
	});

	createEffect(() => {
		location.pathname;
		requestAnimationFrame(updateIndicator);
	});

	return (
		<NavbarContext.Provider value={{ indicatorStyle }}>
			<nav
				ref={navRef}
				class={twMerge("relative flex items-center gap-6", props.class)}
			>
				{props.children}
			</nav>
		</NavbarContext.Provider>
	);
};

const Indicator: Component = () => {
	const context = useContext(NavbarContext);

	if (!context) {
		return null;
	}

	return (
		<span
			aria-hidden="true"
			class="pointer-events-none absolute bottom-0 left-0 h-0.5 rounded-full bg-ctp-sky transition-[transform,width,opacity] duration-300 ease-out"
			style={context.indicatorStyle()}
		/>
	);
};

type NavbarComponent = Component<NavbarProps> & {
	Indicator: Component;
};

export const Navbar = Object.assign(NavbarRoot, {
	Indicator,
}) as NavbarComponent;
