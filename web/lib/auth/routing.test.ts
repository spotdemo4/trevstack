import assert from "node:assert/strict";
import { test } from "node:test";

import { getReturnPath, getSignInPath, isPublicPath } from "./routing.ts";

const origin = "https://stack.example";

void test("public paths normalize case and trailing slashes without matching child routes", () => {
  for (const path of ["/auth", "/AUTH/", "/403", "/403///"]) {
    assert.equal(isPublicPath(path), true, path);
  }
  for (const path of ["/", "", "/numbers", "/metrics", "/auth/other", "/403/other"]) {
    assert.equal(isPublicPath(path), false, path);
  }
});

void test("sign-in round trips the protected path, query, and hash", () => {
  const path = "/numbers?filter=a%26b&sort=desc#results";
  const target = new URL(getSignInPath(path), origin);
  assert.equal(target.pathname, "/auth");
  assert.equal(getReturnPath(target.search, origin), path);
});

void test("return paths reject external destinations and sign-in loops", () => {
  for (const path of [
    "https://other.example",
    "//other.example/path",
    "/\\other.example/path",
    "numbers",
    "/auth",
    "/AUTH///?returnTo=/numbers",
  ]) {
    assert.equal(getReturnPath(`?returnTo=${encodeURIComponent(path)}`, origin), "/", path);
  }
  assert.equal(getReturnPath("", origin), "/");
  assert.equal(getReturnPath("?returnTo=", origin), "/");
});
