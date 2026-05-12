import { Drawer as ArkDrawer } from "@ark-ui/solid/drawer";
import { X } from "lucide-solid";
import { type Component, splitProps } from "solid-js";
import { Portal } from "solid-js/web";
import { twMerge } from "tailwind-merge";

type ContentProps = ArkDrawer.ContentProps & {
  backdropClass?: string;
  positionerClass?: string;
};

export const Root: Component<ArkDrawer.RootProps> = (props) => {
  return <ArkDrawer.Root {...props} />;
};

export const Trigger: Component<ArkDrawer.TriggerProps> = (props) => {
  return (
    <ArkDrawer.Trigger
      {...props}
      class={twMerge(
        "m-2 inline-flex items-center gap-2 self-start rounded-md bg-ctp-surface0 px-3 py-1.5 text-sm font-medium text-ctp-text hover:bg-ctp-surface1",
        props.class,
      )}
    />
  );
};

export const Content: Component<ContentProps> = (props) => {
  const [local, contentProps] = splitProps(props, ["backdropClass", "positionerClass", "class"]);

  return (
    <Portal>
      <ArkDrawer.Backdrop
        class={twMerge(
          "fixed inset-0 z-40 bg-ctp-crust/60 transition-all duration-200 ease-out data-[state=closed]:opacity-0 data-[state=open]:opacity-100 motion-reduce:transition-none",
          local.backdropClass,
        )}
      />
      <ArkDrawer.Positioner
        class={twMerge("fixed inset-y-0 left-0 z-50 flex", local.positionerClass)}
      >
        <ArkDrawer.Content
          {...contentProps}
          class={twMerge(
            "flex h-full w-80 max-w-[85vw] flex-col gap-4 overflow-y-auto bg-ctp-mantle p-4 shadow-xl transition-all duration-300 ease-[cubic-bezier(0.22,1,0.36,1)] data-[state=closed]:-translate-x-full data-[state=closed]:scale-[0.98] data-[state=closed]:opacity-0 data-[state=open]:translate-x-0 data-[state=open]:scale-100 data-[state=open]:opacity-100 motion-reduce:transition-none",
            local.class,
          )}
        />
      </ArkDrawer.Positioner>
    </Portal>
  );
};

export const Title: Component<ArkDrawer.TitleProps> = (props) => {
  return <ArkDrawer.Title {...props} class={twMerge("font-medium", props.class)} />;
};

export const CloseTrigger: Component<ArkDrawer.CloseTriggerProps> = (props) => {
  return (
    <ArkDrawer.CloseTrigger
      {...props}
      class={twMerge(
        "rounded-md p-1 text-ctp-subtext0 hover:bg-ctp-surface0 hover:text-ctp-text",
        props.class,
      )}
    >
      {props.children ?? <X size={16} />}
    </ArkDrawer.CloseTrigger>
  );
};

export default {
  Root,
  Trigger,
  Content,
  Title,
  CloseTrigger,
};
