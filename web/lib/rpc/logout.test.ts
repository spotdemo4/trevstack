import { LoginResponseSchema } from "$gen/auth/v1/login_pb";
import { LogoutResponseSchema } from "$gen/auth/v1/logout_pb";
import { create } from "@bufbuild/protobuf";
import { Code, TrevRpcError } from "@trevrpc/trevrpc-js";
import { Effect } from "effect";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { session } from "../auth/session.ts";
import { AuthClient } from "./clients.ts";
import { logout } from "./logout.ts";

vi.mock("./clients.ts", () => ({ AuthClient: { logout: vi.fn() } }));

const logoutCall = vi.mocked(AuthClient.logout);

beforeEach(() => {
  session.setFromLogin(
    create(LoginResponseSchema, {
      token: "session-token",
      sub: "trev",
      exp: BigInt(Math.floor(Date.now() / 1000) + 3600),
    }),
  );
});

afterEach(() => {
  session.clear();
  localStorage.clear();
  logoutCall.mockReset();
});

test("forgets the session and ends it on the server with its token", () => {
  logoutCall.mockImplementation(() => {
    expect(session.claims()).toBeNull();
    return Effect.succeed(create(LogoutResponseSchema));
  });

  logout();

  expect(session.claims()).toBeNull();
  expect(logoutCall).toHaveBeenCalledExactlyOnceWith(
    {},
    { metadata: { authorization: "Bearer session-token" } },
  );
});

test("stays signed out when the server cannot end the session", async () => {
  logoutCall.mockReturnValue(Effect.fail(new TrevRpcError(Code.Unavailable, "offline")));

  logout();
  await Promise.resolve();

  expect(session.claims()).toBeNull();
});

test("does not call the server without a session", () => {
  session.clear();

  logout();

  expect(logoutCall).not.toHaveBeenCalled();
});
