import { Component } from "solid-js";
import { twMerge } from "tailwind-merge";

type SkeletonProps = {
  class?: string;
};

const Skeleton: Component<SkeletonProps> = (props) => (
  <div class={twMerge("h-4 animate-pulse rounded bg-ctp-surface2", props.class)} />
);

export default Skeleton;
