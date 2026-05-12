import { NumberService } from "$connect/number/v1/service_pb";
import { createClient } from "@connectrpc/connect";

import { transport } from "./transport";

export const NumberClient = createClient(NumberService, transport);
