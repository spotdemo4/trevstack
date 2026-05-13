import { NumberService } from "$connect/number/v1/service_pb";
import { createClient } from "@connectrpc/connect";

import { createEffectClient } from "./effect-client";
import { transport } from "./transport";

export const NumberClient = createEffectClient(
  NumberService,
  createClient(NumberService, transport),
);
