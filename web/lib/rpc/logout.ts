import { bearerMetadata } from "@trevrpc/trevrpc-js";
import { Effect } from "effect";

import { session } from "../auth/session.ts";
import { AuthClient } from "./clients.ts";

/**
 * Signs out: forgets the session at once, so signing out works offline, and
 * asks the server to end it. A session the server could not end still
 * expires. The token goes on the call itself, because clearing the session
 * removes it from the channel's connection metadata.
 */
export function logout(): void {
  const { claims } = session.peek();
  session.clear();
  if (!claims) return;

  let metadata;
  try {
    metadata = bearerMetadata(claims.token);
  } catch {
    // Not a bearer token, which only edited storage holds, so it names no
    // session to end.
    return;
  }
  Effect.runFork(AuthClient.logout({}, { metadata }).pipe(Effect.ignore));
}
