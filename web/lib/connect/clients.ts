import { NumberService } from "$connect/number/v1/service_pb";

import { transport } from "./transport";
import { createSafeClient } from "./wrapper";

export const NumberClient = createSafeClient(NumberService, transport);
