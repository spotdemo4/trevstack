import { createConnectTransport } from "@connectrpc/connect-web";

import { authInterceptor } from "./auth-interceptor";

export const transport = createConnectTransport({
  baseUrl: `${window.location.origin}/grpc`,
  interceptors: [authInterceptor],
  useBinaryFormat: import.meta.env.PROD,
});
