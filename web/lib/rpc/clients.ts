import { AuthService } from "$gen/auth/v1/service_pb";
import { NumberService } from "$gen/number/v1/service_pb";
import { createEffectClient } from "@trevrpc/effect";

import { channel } from "./channel";

export const AuthClient = createEffectClient(AuthService, channel);

export const NumberClient = createEffectClient(NumberService, channel);
