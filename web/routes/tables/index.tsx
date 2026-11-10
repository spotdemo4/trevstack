import { Tabs } from "$lib/tabs";
import type { Component } from "solid-js";

import { createPositions, Positions } from "./positions";
import { createStreams, Streams } from "./streams";

const Tables: Component = () => {
  // Created here rather than in each tab, so filters and results survive switching tabs.
  const positions = createPositions();
  const streams = createStreams();

  return (
    <Tabs.Root defaultValue="positions" class="gap-0 min-[900px]:h-body">
      <div class="border-b border-line bg-surface px-4 py-2">
        <Tabs.List aria-label="Tables">
          <Tabs.Trigger value="positions">Positions</Tabs.Trigger>
          <Tabs.Trigger value="streams">Streams</Tabs.Trigger>
        </Tabs.List>
      </div>
      <Tabs.Content value="positions" class="min-h-0 flex-1">
        <Positions state={positions} />
      </Tabs.Content>
      <Tabs.Content value="streams" class="min-h-0 flex-1">
        <Streams state={streams} />
      </Tabs.Content>
    </Tabs.Root>
  );
};

export default Tables;
