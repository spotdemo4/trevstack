import { createChannel } from "@trevrpc/trevrpc-js";

import { session } from "../auth/session.ts";
import { authInterceptor, sessionMetadata } from "./auth-interceptor.ts";

/** The server's TrevRPC endpoint; the page's own server unless VITE_RPC_URL names another. */
const rpcUrl = import.meta.env.VITE_RPC_URL || `${window.location.origin}/trevrpc`;

/** The channel every client shares, which starts connecting when the app loads. */
export const channel = createChannel(rpcUrl, {
  // A server with a development certificate publishes the hashes browsers
  // trust it by; a 404 means the browser already trusts its certificate.
  serverCertificateHashesUrl: "/certificate-hashes",
  interceptors: [authInterceptor],
});

// Every call carries the current session's token as connection metadata. The
// session updates it synchronously, so calls started after a sign-in or
// sign-out never carry the old token.
session.subscribe((claims) => {
  try {
    channel.setConnectionMetadata(sessionMetadata(claims));
  } catch {
    // A token too large to send, which only edited storage holds: send none,
    // so the next protected call fails as unauthenticated and signs out.
    channel.setConnectionMetadata();
  }
});
