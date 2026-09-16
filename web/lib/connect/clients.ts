import { AuthService } from "$connect/auth/v1/service_pb";
import { NumberService } from "$connect/number/v1/service_pb";
import { createClient } from "@connectrpc/connect";

import { createEffectClient } from "./effect-client";
import { transport } from "./transport";

export const AuthClient = createEffectClient(AuthService, createClient(AuthService, transport));

export const NumberClient = createEffectClient(
  NumberService,
  createClient(NumberService, transport),
);
