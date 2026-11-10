import { AuthService } from "$gen/auth/v1/service_pb";
import { CursorService } from "$gen/cursor/v1/service_pb";
import { createEffectClient } from "@trevrpc/effect";
import { createRequestValidator } from "@trevrpc/protovalidate";

import { channel } from "./channel";

// Requests that break the buf.validate rules in their protos fail with InvalidArgument before
// they are sent, as the server would reject them.
const validator = createRequestValidator();

export const AuthClient = createEffectClient(AuthService, channel, { validator });
export const CursorClient = createEffectClient(CursorService, channel, { validator });
