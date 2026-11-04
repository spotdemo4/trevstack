import { createChannel } from "@trevrpc/trevrpc-js";

import { authInterceptor } from "./auth-interceptor.ts";

/** The server's TrevRPC endpoint; the page's own server unless VITE_RPC_URL names another. */
const rpcUrl = import.meta.env.VITE_RPC_URL || `${window.location.origin}/trevrpc`;

/** The channel every client shares, which starts connecting when the app loads. */
export const channel = createChannel(rpcUrl, {
  // A server with a development certificate publishes the hashes browsers
  // trust it by; a 404 means the browser already trusts its certificate.
  serverCertificateHashesUrl: "/certificate-hashes",
  interceptors: [authInterceptor],
});
