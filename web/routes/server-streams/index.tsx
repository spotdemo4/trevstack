import { createCursorStore, CursorPad, LiveStatus } from "$lib/cursor-pad";
import { createLiveStream } from "$lib/effect";
import { PageHeader } from "$lib/page-header";
import { CursorClient } from "$lib/rpc";
import type { Component } from "solid-js";

const ServerStreams: Component = () => {
  const cursors = createCursorStore();
  const watch = createLiveStream(
    () => {
      // A new stream starts with the open streams' positions, so forget the old ones.
      cursors.clear();
      return CursorClient.watch({});
    },
    (response) => cursors.apply(response.event),
  );

  return (
    <div class="mx-auto flex max-w-5xl flex-col gap-4 p-4">
      <PageHeader
        title="Server streams"
        kind="Server streaming"
        rpc="rpc Watch(WatchRequest) returns (stream WatchResponse)"
      >
        One request, then a response for every cursor that moves, for as long as the page is open.
        Send positions from the other pages, another browser, or the command-line client to see them
        here.
      </PageHeader>
      <LiveStatus stream={watch} />
      <CursorPad label="Live cursors" cursors={cursors.cursors} />
      <p class="text-sm text-foreground-subtle">
        {cursors.cursors.length === 1 ? "1 cursor" : `${cursors.cursors.length} cursors`}
      </p>
    </div>
  );
};

export default ServerStreams;
