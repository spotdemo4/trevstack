import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

import { getSignInPath, normalizePath } from "../auth/routing.ts";
import { session } from "../auth/session.ts";

export const authInterceptor: Interceptor = (next) => async (request) => {
  if (
    request.service.typeName === "auth.v1.AuthService" &&
    ["Login", "Signup", "Logout"].includes(request.method.name)
  ) {
    return next(request);
  }

  const revision = session.revision();
  try {
    const response = await next(request);
    if (!response.stream) {
      return response;
    }
    return {
      ...response,
      message: redirectingIterable(response.message, revision),
    };
  } catch (error) {
    redirectForAuthError(error, revision);
    throw error;
  }
};

async function* redirectingIterable<T>(
  source: AsyncIterable<T>,
  revision: number,
): AsyncIterable<T> {
  try {
    yield* source;
  } catch (error) {
    redirectForAuthError(error, revision);
    throw error;
  }
}

function redirectForAuthError(error: unknown, revision: number) {
  if (revision !== session.revision()) return;

  const connectError = ConnectError.from(error);
  if (connectError.code === Code.Unauthenticated) {
    const { pathname, search, hash } = window.location;
    const target = getSignInPath(`${pathname}${search}${hash}`);
    session.clear();
    if (normalizePath(pathname) !== "/auth") {
      window.location.replace(target);
    }
    return;
  }

  if (
    connectError.code === Code.PermissionDenied &&
    normalizePath(window.location.pathname) !== "/403"
  ) {
    window.location.replace("/403");
  }
}
