import assert from "node:assert/strict";
import { afterEach, beforeEach, test } from "node:test";

import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

import { session } from "../auth/session.ts";
import { authInterceptor } from "./auth-interceptor.ts";

type Request = Parameters<ReturnType<Interceptor>>[0];
type Response = Awaited<ReturnType<ReturnType<Interceptor>>>;

let redirects: string[];
let originalWindow: PropertyDescriptor | undefined;

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
  originalWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
  const values = new Map<string, string>();
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: {
      addEventListener() {},
      removeEventListener() {},
      location: {
        pathname: "/numbers",
        search: "?sort=desc",
        hash: "#results",
        replace: (url: string) => redirects.push(url),
      },
      localStorage: {
        getItem: (key: string) => values.get(key) ?? null,
        setItem: (key: string, value: string) => values.set(key, value),
        removeItem: (key: string) => values.delete(key),
      },
    },
  });
  session.setFromLogin(loginMetadata());
});

afterEach(() => {
  session.clear();
  if (originalWindow) Object.defineProperty(globalThis, "window", originalWindow);
  else Reflect.deleteProperty(globalThis, "window");
});

void test("unary unauthenticated failures clear the session before preserving the return path", async () => {
  const error = new ConnectError("session expired", Code.Unauthenticated);
  window.location.replace = (url: string) => {
    assert.equal(session.claims(), null);
    redirects.push(url);
  };
  const invoke = authInterceptor(async () => {
    throw error;
  });

  await assert.rejects(invoke(request()), (caught) => caught === error);
  assert.equal(session.claims(), null);
  assert.deepEqual(redirects, ["/auth?returnTo=%2Fnumbers%3Fsort%3Ddesc%23results"]);
});

void test("unauthenticated failures on the auth page do not redirect in a loop", async () => {
  window.location.pathname = "/AUTH/";
  const invoke = authInterceptor(async () => {
    throw new ConnectError("missing session", Code.Unauthenticated);
  });

  await assert.rejects(invoke(request()));
  assert.equal(session.claims(), null);
  assert.deepEqual(redirects, []);
});

void test("permission denied redirects without clearing the hint", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("invalid session", Code.PermissionDenied);
  });
  await assert.rejects(invoke(request()));
  assert.equal(session.claims()?.sub, "trev");
  assert.deepEqual(redirects, ["/403"]);
});

void test("network failures leave the session and current page alone", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("offline", Code.Unavailable);
  });
  await assert.rejects(invoke(request()));
  assert.equal(session.claims()?.sub, "trev");
  assert.deepEqual(redirects, []);
});

void test("public auth failures are left to the form or logout control", async () => {
  const invoke = authInterceptor(async () => {
    throw new ConnectError("invalid credentials", Code.Unauthenticated);
  });
  for (const method of ["Login", "Signup", "Logout"]) {
    await assert.rejects(invoke(request("auth.v1.AuthService", method)));
    assert.equal(session.claims()?.sub, "trev");
    assert.deepEqual(redirects, []);
  }
});

void test("old unary failures cannot invalidate a newer login or redirect after logout", async () => {
  for (const changeSession of [
    () => session.setFromLogin(loginMetadata("new-user")),
    () => session.clear(),
  ]) {
    const invoke = authInterceptor(async () => {
      changeSession();
      throw new ConnectError("old session expired", Code.Unauthenticated);
    });
    await assert.rejects(invoke(request()));
    assert.deepEqual(redirects, []);
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

void test("streaming failures clear the session and preserve the underlying error", async () => {
  const response = await failingStream();
  assert.equal(response.stream, true);
  if (!response.stream) return;

  await assert.rejects(async () => {
    for await (const _message of response.message) {
      assert.fail("the failing stream should not yield a message");
    }
  }, ConnectError);
  assert.equal(session.claims(), null);
  assert.equal(redirects.length, 1);
});

void test("old streaming failures cannot invalidate a newer session", async () => {
  const response = await failingStream();
  session.setFromLogin(loginMetadata("new-user"));
  if (!response.stream) assert.fail("expected a stream");

  await assert.rejects(async () => {
    for await (const _message of response.message) {
      assert.fail("the failing stream should not yield a message");
    }
  });
  assert.equal(session.claims()?.sub, "new-user");
  assert.deepEqual(redirects, []);
});
