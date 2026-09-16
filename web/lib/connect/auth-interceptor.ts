import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

export const authInterceptor: Interceptor = (next) => async (request) => {
  try {
    const response = await next(request);
    if (!response.stream) {
      return response;
    }
    return {
      ...response,
      message: redirectingIterable(response.message),
    };
  } catch (error) {
    redirectForAuthError(error);
    throw error;
  }
};

async function* redirectingIterable<T>(source: AsyncIterable<T>): AsyncIterable<T> {
  try {
    yield* source;
  } catch (error) {
    redirectForAuthError(error);
    throw error;
  }
}

function redirectForAuthError(error: unknown) {
  const connectError = ConnectError.from(error);
  if (connectError.code === Code.Unauthenticated) {
    if (window.location.pathname === "/auth") {
      return;
    }
    const returnTo = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    window.location.replace(`/auth?returnTo=${encodeURIComponent(returnTo)}`);
    return;
  }

  if (connectError.code === Code.PermissionDenied && window.location.pathname !== "/403") {
    window.location.replace("/403");
  }
}
