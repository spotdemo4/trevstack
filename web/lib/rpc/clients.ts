import { AuthServiceClient, AuthServiceService } from "$gen/auth/v1/service.trevrpc";
import { NumberServiceClient, NumberServiceService } from "$gen/number/v1/service.trevrpc";

import { channel } from "./channel";
import { createEffectClient } from "./effect-client";

export const AuthClient = createEffectClient(AuthServiceService, new AuthServiceClient(channel));

export const NumberClient = createEffectClient(
  NumberServiceService,
  new NumberServiceClient(channel),
);
