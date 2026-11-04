import { Code, type CallInterceptor, TrevRpcError } from "@trevrpc/trevrpc-js";

import { getSignInPath, normalizePath } from "../auth/routing.ts";
import { session } from "../auth/session.ts";

/** Methods that do not need a session, whose failures their forms report. */
const PUBLIC_METHODS = new Set(["auth.v1.AuthService/Login", "auth.v1.AuthService/Signup"]);

export function isPublicMethod(service: string, method: string): boolean {
  return PUBLIC_METHODS.has(`${service}/${method}`);
}

/**
 * Sends the session token with every call to a protected method, and handles
 * the call's authentication failure when it ends.
 */
export const authInterceptor: CallInterceptor = (call) => {
  if (isPublicMethod(call.service, call.name)) return;

  // Calls start inside effects, so read the session without tracking it.
  const { claims, revision } = session.peek();
  if (claims) call.metadata.authorization = `Bearer ${claims.token}`;
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
