import { createConnectTransport } from "@connectrpc/connect-web";

export const transport = createConnectTransport({
  baseUrl: `${window.location.origin}/grpc`,
  useBinaryFormat: import.meta.env.PROD,
});
