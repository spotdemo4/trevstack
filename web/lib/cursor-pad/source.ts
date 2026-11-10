import { Source } from "$gen/cursor/v1/cursor_pb";

/** The RPC that recorded a stream, named for people. */
export const sourceLabel = (source: Source): string => {
  switch (source) {
    case Source.SEND:
      return "Send";
    case Source.TRACK:
      return "Track";
    case Source.SHARE:
      return "Share";
    default:
      return "Unknown";
  }
};

/** Every source, as select items. */
export const sourceItems = [Source.SEND, Source.TRACK, Source.SHARE].map((source) => ({
  label: sourceLabel(source),
  value: source,
}));
