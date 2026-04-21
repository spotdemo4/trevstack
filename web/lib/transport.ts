import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { GreetService } from "$connect/greet/v1/greet_pb";

const transport = createConnectTransport({
	baseUrl: `${window.location.origin}/grpc`,
});

export const GreetClient = createClient(GreetService, transport);
