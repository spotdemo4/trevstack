import { public$ } from "$gen/auth/v1/options_pb";
import { type DescMethod, getOption } from "@bufbuild/protobuf";
import { bearerMetadata, Code, type CallInterceptor, TrevRpcError } from "@trevrpc/trevrpc-js";

import { getSignInPath, normalizePath } from "../auth/routing.ts";
import { session, type SessionClaims } from "../auth/session.ts";

/**
 * Whether a method needs no session, which its `(auth.v1.public)` option says, as the server reads
 * it. The callers of public methods handle their failures: the forms report sign-in and sign-up
 * failures, and signing out ignores them.
 */
export function isPublicMethod(method: DescMethod): boolean {
  return getOption(method, public$);
}

/**
 * The channel's connection metadata for a session: its token, which every call
 * carries and which crosses each connection once, or none without a session.
 * Throws for a token that is not a bearer token, which only edited storage holds.
 */
export function sessionMetadata(claims: SessionClaims | null): Record<string, string> | undefined {
  return claims ? bearerMetadata(claims.token) : undefined;
}

/**
 * Handles the authentication failure of a call to a protected method when it
 * ends. The session token comes from the channel's connection metadata.
 */
export const authInterceptor: CallInterceptor = (call) => {
  if (isPublicMethod(call.descriptor)) return;

  // Calls start inside effects, so read the session without tracking it.
  const { revision } = session.peek();
  return ({ error }) => handleAuthError(error, revision);
};

/**
 * Signs out and redirects to the sign-in page when a call fails because the
 * session is missing or expired, and shows the forbidden page when it is
 * invalid. Failures of calls started before the session changed are ignored,
 * so an old call cannot sign out a newer session, and so are failures without
 * a session, whose pages the app already sends to the sign-in page.
 */
export function handleAuthError(error: unknown, revision: number): void {
  const current = session.peek();
  if (revision !== current.revision || current.claims === null) return;
  if (!(error instanceof TrevRpcError)) return;

  if (error.code === Code.Unauthenticated) {
    const { pathname, search, hash } = window.location;
    const target = getSignInPath(`${pathname}${search}${hash}`);
    session.clear();
    if (normalizePath(pathname) !== "/auth") {
      window.location.replace(target);
    }
    return;
  }

  if (error.code === Code.PermissionDenied && normalizePath(window.location.pathname) !== "/403") {
    window.location.replace("/403");
  }
}
