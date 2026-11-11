import { LoginResponseSchema } from "$gen/auth/v1/login_pb";
import { AuthService } from "$gen/auth/v1/service_pb";
import { CursorService } from "$gen/cursor/v1/service_pb";
import { create, type DescMethod } from "@bufbuild/protobuf";
import { type CallFinish, Code, type InterceptedCall, TrevRpcError } from "@trevrpc/trevrpc-js";
import { afterEach, beforeEach, expect, type MockInstance, test, vi } from "vitest";

import { session } from "../auth/session.ts";
import {
  authInterceptor,
  handleAuthError,
  isPublicMethod,
  sessionMetadata,
} from "./auth-interceptor.ts";

let redirects: string[];
let replace: MockInstance<Location["replace"]>;

function loginMetadata(sub = "trev") {
  return create(LoginResponseSchema, {
    token: `${sub}.token`,
    sub,
    exp: BigInt(Math.floor(Date.now() / 1000) + 3600),
  });
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

test("unauthenticated failures clear the session before preserving the return path", () => {
  replace.mockImplementation((url) => {
    expect(session.claims()).toBeNull();
    redirects.push(String(url));
  });

  handleAuthError(new TrevRpcError(Code.Unauthenticated, "session expired"), session.revision());

  expect(session.claims()).toBeNull();
  expect(redirects).toEqual(["/auth?returnTo=%2Fnumbers%3Fsort%3Ddesc%23results"]);
});

test("unauthenticated failures on the auth page do not redirect in a loop", () => {
  history.replaceState(null, "", "/AUTH/");

  handleAuthError(new TrevRpcError(Code.Unauthenticated, "missing session"), session.revision());

  expect(session.claims()).toBeNull();
  expect(redirects).toEqual([]);
});

test("permission denied redirects without clearing the hint", () => {
  handleAuthError(new TrevRpcError(Code.PermissionDenied, "invalid session"), session.revision());

  expect(session.claims()?.sub).toBe("trev");
  expect(redirects).toEqual(["/403"]);
});

test("network and other failures leave the session and current page alone", () => {
  for (const error of [
    new TrevRpcError(Code.Unavailable, "offline"),
    new TrevRpcError(Code.InvalidArgument, "name is required"),
    new Error("not a status"),
  ]) {
    handleAuthError(error, session.revision());
  }

  expect(session.claims()?.sub).toBe("trev");
  expect(redirects).toEqual([]);
});

test("failures of calls started before the session changed are ignored", () => {
  for (const changeSession of [
    () => session.setFromLogin(loginMetadata("new-user")),
    () => session.clear(),
  ]) {
    const revision = session.revision();
    changeSession();
    handleAuthError(new TrevRpcError(Code.Unauthenticated, "old session expired"), revision);
    expect(redirects).toEqual([]);
  }
});

test("failures without a session leave navigation to the app", () => {
  session.clear();

  handleAuthError(
    new TrevRpcError(Code.Unauthenticated, "authentication required"),
    session.revision(),
  );

  expect(redirects).toEqual([]);
});

function call(descriptor: DescMethod = CursorService.method.list): InterceptedCall {
  const service = descriptor.parent.typeName;
  return {
    method: `/${service}/${descriptor.name}`,
    service,
    name: descriptor.name,
    kind: "serverStreaming",
    descriptor,
    metadata: {},
    timeoutMs: undefined,
    signal: undefined,
  };
}

async function intercept(intercepted: InterceptedCall) {
  return (await authInterceptor(intercepted)) as CallFinish | undefined;
}

test("protected calls handle their authentication failures, leaving the token to the channel", async () => {
  const intercepted = call();
  const finish = await intercept(intercepted);

  expect(intercepted.metadata).toEqual({});
  finish?.({
    code: Code.Unauthenticated,
    error: new TrevRpcError(Code.Unauthenticated, "expired"),
  });
  expect(session.claims()).toBeNull();
  expect(redirects).toHaveLength(1);
});

test("protected calls started under an older session are ignored when they end", async () => {
  const finish = await intercept(call());
  session.setFromLogin(loginMetadata("new-user"));

  finish?.({
    code: Code.Unauthenticated,
    error: new TrevRpcError(Code.Unauthenticated, "expired"),
  });
  expect(session.claims()?.sub).toBe("new-user");
  expect(redirects).toEqual([]);
});

test("public calls leave their failures to the forms", async () => {
  const intercepted = call(AuthService.method.login);

  expect(await intercept(intercepted)).toBeUndefined();
  expect(intercepted.metadata).toEqual({});
});

test("connection metadata carries the session token, or nothing without a session", () => {
  expect(sessionMetadata(session.peek().claims)).toEqual({ authorization: "Bearer trev.token" });
  expect(sessionMetadata(null)).toBeUndefined();
});

test("connection metadata rejects a token that is not a bearer token", () => {
  const claims = session.peek().claims;
  if (!claims) throw new Error("no session");
  expect(() => sessionMetadata({ ...claims, token: "not a token" })).toThrow(
    expect.objectContaining({ code: Code.InvalidArgument }),
  );
});

test("only login, logout and signup are public", () => {
  expect(isPublicMethod(AuthService.method.login)).toBe(true);
  expect(isPublicMethod(AuthService.method.logout)).toBe(true);
  expect(isPublicMethod(AuthService.method.signup)).toBe(true);
  for (const method of CursorService.methods) {
    expect(isPublicMethod(method)).toBe(false);
  }
});
