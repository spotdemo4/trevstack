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

  Effect.runFork(
    AuthClient.logout({}, { metadata: { authorization: `Bearer ${claims.token}` } }).pipe(
      Effect.ignore,
    ),
  );
}
