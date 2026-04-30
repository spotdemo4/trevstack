import { Splitter } from "@ark-ui/solid/splitter";
import { type Component, mergeProps } from "solid-js";
import styles from "./splitter.module.css";

export const Root: Component<Splitter.RootProps> = (props) => {
	return <Splitter.Root {...props} />;
};

export const Panel: Component<Splitter.PanelProps> = (props) => {
	return <Splitter.Panel {...props} />;
};

export const ResizeTrigger: Component<Splitter.ResizeTriggerProps> = (
	props,
) => {
	const finalProps = mergeProps(
		{ class: styles.ResizeTrigger, "aria-label": "Resize" },
		props,
	);

	return (
		<Splitter.ResizeTrigger {...finalProps}>
			<Splitter.ResizeTriggerIndicator class={styles.ResizeTriggerIndicator} />
		</Splitter.ResizeTrigger>
	);
};

export default {
	Root,
	Panel,
	ResizeTrigger,
};
