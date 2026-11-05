import { AuthServiceClient, AuthServiceService } from "$gen/auth/v1/service.trevrpc";
import { NumberServiceClient, NumberServiceService } from "$gen/number/v1/service.trevrpc";
import { createEffectClient } from "@trevrpc/effect";

import { channel } from "./channel";

export const AuthClient = createEffectClient(AuthServiceService, new AuthServiceClient(channel));

export const NumberClient = createEffectClient(
  NumberServiceService,
  new NumberServiceClient(channel),
);
