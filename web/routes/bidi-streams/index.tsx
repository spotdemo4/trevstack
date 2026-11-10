import type { ShareRequestSchema } from "$gen/cursor/v1/share_pb";
import { session } from "$lib/auth";
import { createCursorStore, CursorPad, LiveStatus, type PadPoint } from "$lib/cursor-pad";
import { createLiveStream, createPushStream, type PushStream } from "$lib/effect";
import { PageHeader } from "$lib/page-header";
import { CursorClient } from "$lib/rpc";
import type { MessageInitShape } from "@bufbuild/protobuf";
import type { Component } from "solid-js";

/** The key of your own cursor, which no stream ID can collide with. */
const OWN_CURSOR = "own";

const BidiStreams: Component = () => {
  const cursors = createCursorStore();
  let requests: PushStream<MessageInitShape<typeof ShareRequestSchema>> | undefined;

  const share = createLiveStream(
    () => {
      cursors.clear();
      requests = createPushStream();
      return CursorClient.share(requests.stream);
    },
    (response) => cursors.apply(response.event),
  );

  const move = (point: PadPoint) => {
    requests?.push(point);
    // The server never sends your own stream back, so draw your cursor here.
    cursors.move(OWN_CURSOR, session.claims()?.sub ?? "you", point);
  };

  return (
    <div class="mx-auto flex max-w-5xl flex-col gap-4 p-4">
      <PageHeader
        title="Bidi streams"
        kind="Bidirectional streaming"
        rpc="rpc Share(stream ShareRequest) returns (stream ShareResponse)"
      >
        One call carries your cursor to the server and everyone else's back, both at once. Move over
        the pad, then open this page in another window or as another user.
      </PageHeader>
      <LiveStatus stream={share} />
      <CursorPad
        label="Shared cursors"
        cursors={cursors.cursors}
        onMove={move}
        onLeave={() => cursors.remove(OWN_CURSOR)}
        class="cursor-none"
      />
    </div>
  );
};

export default BidiStreams;
