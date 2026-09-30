import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";
import { afterEach, beforeEach, expect, type MockInstance, test, vi } from "vitest";

import { session } from "../auth/session.ts";
import { authInterceptor } from "./auth-interceptor.ts";

type Request = Parameters<ReturnType<Interceptor>>[0];
type Response = Awaited<ReturnType<ReturnType<Interceptor>>>;

let redirects: string[];
let replace: MockInstance<Location["replace"]>;

function loginMetadata(sub = "trev") {
  return { sub, exp: BigInt(Math.floor(Date.now() / 1000) + 3600) };
}

function request(service = "number.v1.NumberService", method = "List"): Request {
  return {
    stream: false,
    service: { typeName: service },
    method: { name: method },
  } as unknown as Request;
}

beforeEach(() => {
  redirects = [];
  history.replaceState(null, "", "/numbers?sort=desc#results");
  replace = vi.spyOn(window.location, "replace").mockImplementation((url) => {
    redirects.push(String(url));
  });
  session.setFromLogin(loginMetadata());
});

afterEach(() => {
  session.clear();
  localStorage.clear();
});

test("unary unauthenticated failures clear the session before preserving the return path", async () => {
  const error = new ConnectError("session expired", Code.Unauthenticated);
  replace.mockImplementation((url) => {
    expect(session.claims()).toBeNull();
    redirects.push(String(url));
  });
  const invoke = authInterceptor(async () => {
    throw error;
  });

  await expect(invoke(request())).rejects.toBe(error);
  expect(session.claims()).toBeNull();
  expect(redirects).toEqual(["/auth?returnTo=%2Fnumbers%3Fsort%3Ddesc%23results"]);
});

test("unauthenticated failures on the auth page do not redirect in a loop", async () => {
  history.replaceState(null, "", "/AUTH/");
  const invoke = authInterceptor(async () => {
    throw new ConnectError("missing session", Code.Unauthenticated);
  });

  await expect(invoke(request())).rejects.toThrow(ConnectError);
  expect(session.claims()).toBeNull();
  expect(redirects).toEqual([]);
});

test("permission denied redirects without clearing the hint", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("invalid session", Code.PermissionDenied);
  });
  await expect(invoke(request())).rejects.toThrow(ConnectError);
  expect(session.claims()?.sub).toBe("trev");
  expect(redirects).toEqual(["/403"]);
});

test("network failures leave the session and current page alone", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("offline", Code.Unavailable);
  });
  await expect(invoke(request())).rejects.toThrow(ConnectError);
  expect(session.claims()?.sub).toBe("trev");
  expect(redirects).toEqual([]);
});

test("public auth failures are left to the form or logout control", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("invalid credentials", Code.Unauthenticated);
  });
  for (const method of ["Login", "Signup", "Logout"]) {
    await expect(invoke(request("auth.v1.AuthService", method))).rejects.toThrow(ConnectError);
    expect(session.claims()?.sub).toBe("trev");
    expect(redirects).toEqual([]);
  }
});

test("old unary failures cannot invalidate a newer login or redirect after logout", async () => {
  for (const changeSession of [
    () => session.setFromLogin(loginMetadata("new-user")),
    () => session.clear(),
  ]) {
    const invoke = authInterceptor(async () => {
      changeSession();
      throw new ConnectError("old session expired", Code.Unauthenticated);
    });
    await expect(invoke(request())).rejects.toThrow(ConnectError);
    expect(redirects).toEqual([]);
  }
});

function failingStream() {
  const error = new ConnectError("stream session expired", Code.Unauthenticated);
  const message: AsyncIterable<never> = {
    [Symbol.asyncIterator]() {
      return { next: () => Promise.reject(error) };
    },
  };
  const invoke = authInterceptor(async () => ({ stream: true, message }) as Response);
  return invoke(request());
}

test("streaming failures clear the session and preserve the underlying error", async () => {
  const response = await failingStream();
  expect(response.stream).toBe(true);
  if (!response.stream) return;

  await expect(async () => {
    for await (const _message of response.message) {
      expect.unreachable("the failing stream should not yield a message");
    }
  }).rejects.toThrow(ConnectError);
  expect(session.claims()).toBeNull();
  expect(redirects).toHaveLength(1);
});

test("old streaming failures cannot invalidate a newer session", async () => {
  const response = await failingStream();
  session.setFromLogin(loginMetadata("new-user"));
  if (!response.stream) expect.unreachable("expected a stream");

  await expect(async () => {
    for await (const _message of response.message) {
      expect.unreachable("the failing stream should not yield a message");
    }
  }).rejects.toThrow(ConnectError);
  expect(session.claims()?.sub).toBe("new-user");
  expect(redirects).toEqual([]);
});
