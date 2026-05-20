import { Splitter as ArkSplitter } from "@ark-ui/solid/splitter";
import { type Component, mergeProps } from "solid-js";

import styles from "./splitter.module.css";

export const Root: Component<ArkSplitter.RootProps> = (props) => {
  return <ArkSplitter.Root {...props} />;
};

export const Panel: Component<ArkSplitter.PanelProps> = (props) => {
  return <ArkSplitter.Panel {...props} />;
};

export const ResizeTrigger: Component<ArkSplitter.ResizeTriggerProps> = (props) => {
  const finalProps = mergeProps({ class: styles.ResizeTrigger, "aria-label": "Resize" }, props);

  return (
    <ArkSplitter.ResizeTrigger {...finalProps}>
      <ArkSplitter.ResizeTriggerIndicator class={styles.ResizeTriggerIndicator} />
    </ArkSplitter.ResizeTrigger>
  );
};

export const Splitter = {
  Root,
  Panel,
  ResizeTrigger,
};
