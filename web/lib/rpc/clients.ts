import { AuthService } from "$gen/auth/v1/service_pb";
import { AuthServiceCodecs } from "$gen/auth/v1/service_trevrpc";
import { CursorService } from "$gen/cursor/v1/service_pb";
import { CursorServiceCodecs } from "$gen/cursor/v1/service_trevrpc";
import { createEffectClient } from "@trevrpc/effect";
import { createRequestValidator } from "@trevrpc/protovalidate";

import { channel } from "./channel";

// Requests that break the buf.validate rules in their protos fail with InvalidArgument before
// they are sent, as the server would reject them.
const validator = createRequestValidator();

// The generated codecs encode and decode messages to the same bytes as protobuf-es, faster.
export const AuthClient = createEffectClient(AuthService, channel, {
  validator,
  codecs: AuthServiceCodecs,
});
export const CursorClient = createEffectClient(CursorService, channel, {
  validator,
  codecs: CursorServiceCodecs,
});
