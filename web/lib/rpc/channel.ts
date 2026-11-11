import { createChannel } from "@trevrpc/trevrpc-js";

import { session } from "../auth/session.ts";
import { authInterceptor, sessionMetadata } from "./auth-interceptor.ts";

/** The server's TrevRPC endpoint; the page's own server unless VITE_RPC_URL names another. */
const rpcUrl = import.meta.env.VITE_RPC_URL || `${window.location.origin}/trevrpc`;

/** The channel every client shares, which starts connecting when the app loads. */
export const channel = createChannel(rpcUrl, {
  // A server with a development certificate publishes the hashes browsers
  // trust it by at its endpoint; a 404 means the browser already trusts its
  // certificate. They are read from the page's own server, which is the
  // TrevRPC server in production, and proxies them in development: a
  // browser would not fetch them from a server whose certificate it does
  // not trust.
  serverCertificateHashesUrl: "/trevrpc/certificate-hashes",
  interceptors: [authInterceptor],
});

// Every call carries the current session's token as connection metadata. The
// session updates it synchronously, so calls started after a sign-in or
// sign-out never carry the old token, and calls in flight keep the token they
// started with, so a call's failure is always about the session whose revision
// the auth interceptor recorded for it.
session.subscribe((claims) => {
  try {
    channel.setConnectionMetadata(sessionMetadata(claims));
  } catch {
    // A token that is too large to send or not a bearer token, which only
    // edited storage holds: send none, so the next protected call fails as
    // unauthenticated and signs out.
    channel.setConnectionMetadata();
  }
});
