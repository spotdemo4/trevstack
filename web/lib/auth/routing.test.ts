import { expect, test } from "vitest";

import { getReturnPath, getSignInPath, isPublicPath } from "./routing.ts";

const origin = "https://stack.example";

test("public paths normalize case and trailing slashes without matching child routes", () => {
  for (const path of ["/auth", "/AUTH/", "/403", "/403///"]) {
    expect(isPublicPath(path), path).toBe(true);
  }
  for (const path of ["/", "", "/numbers", "/metrics", "/auth/other", "/403/other"]) {
    expect(isPublicPath(path), path).toBe(false);
  }
});

test("sign-in round trips the protected path, query, and hash", () => {
  const path = "/numbers?filter=a%26b&sort=desc#results";
  const target = new URL(getSignInPath(path), origin);
  expect(target.pathname).toBe("/auth");
  expect(getReturnPath(target.search, origin)).toBe(path);
});

test("return paths reject external destinations and sign-in loops", () => {
  for (const path of [
    "https://other.example",
    "//other.example/path",
    "/\\other.example/path",
    "numbers",
    "/auth",
    "/AUTH///?returnTo=/numbers",
  ]) {
    expect(getReturnPath(`?returnTo=${encodeURIComponent(path)}`, origin), path).toBe("/");
  }
  expect(getReturnPath("", origin)).toBe("/");
  expect(getReturnPath("?returnTo=", origin)).toBe("/");
});
