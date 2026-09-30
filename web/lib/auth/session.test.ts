import { createRoot, onCleanup, untrack } from "solid-js";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { SESSION_STORAGE_KEY, session, type SessionClaims } from "./session.ts";

const NOW = 1_700_000_000_000;

let cleanup: (() => void) | undefined;

// Replaces localStorage with one whose `method` throws, like a browser with
// site data blocked. Calls compose, and unstubGlobals restores the original.
function breakStorage(method: "getItem" | "setItem" | "removeItem") {
  const current = localStorage;
  vi.stubGlobal("localStorage", {
    getItem: (key: string) => current.getItem(key),
    setItem: (key: string, value: string) => current.setItem(key, value),
    removeItem: (key: string) => current.removeItem(key),
    [method]: () => {
      throw new Error("storage unavailable");
    },
  });
}

function stored() {
  return localStorage.getItem(SESSION_STORAGE_KEY);
}

function store(value: unknown) {
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(value));
}

function nowInSeconds() {
  return Math.floor(Date.now() / 1000);
}

function validClaims(overrides: Partial<SessionClaims> = {}): SessionClaims {
  return {
    sub: "user-1",
    exp: nowInSeconds() + 60,
    ...overrides,
  };
}

function loginMetadata(claims: SessionClaims = validClaims()) {
  return { sub: claims.sub, exp: BigInt(claims.exp) };
}

// Moves the clock without firing timers, like a suspended tab being resumed.
function sleep(ms: number) {
  vi.setSystemTime(Date.now() + ms);
}

beforeEach(() => {
  vi.useFakeTimers({ now: NOW });
});

afterEach(() => {
  cleanup?.();
  cleanup = undefined;
  vi.unstubAllGlobals();
  session.clear();
  localStorage.clear();
  vi.useRealTimers();
});

test("accepts explicit login metadata, persists only the exact allowlist, and ignores extras", () => {
  const claims = validClaims({ sub: "用户/é" });
  cleanup = session.initialize();
  const response = {
    ...loginMetadata(claims),
    jwt: "ignored.jwt.value",
    arbitrary: { ignored: true },
  };
  session.setFromLogin(response);

  expect(session.claims()).toEqual(claims);
  expect(stored()).toBe(JSON.stringify(claims));
  expect(Object.keys(JSON.parse(stored()!))).toEqual(["sub", "exp"]);
});

test("rejects invalid login metadata before changing the session", () => {
  cleanup = session.initialize();
  const now = Date.now();
  const invalidMetadata: unknown[] = [
    null,
    undefined,
    [],
    { exp: BigInt(now / 1000 + 60) },
    { sub: "user-1" },
    { sub: "user-1", exp: "future" },
    { sub: "user-1", exp: now / 1000 + 60 },
    { sub: "user-1", exp: BigInt(now / 1000) },
    { sub: "user-1", exp: BigInt(now / 1000 - 1) },
    { sub: "", exp: BigInt(nowInSeconds() + 1) },
    { sub: "user-1", exp: 0n },
    { sub: "user-1", exp: -1n },
    { sub: "user-1", exp: BigInt(Number.MAX_SAFE_INTEGER) + 1n },
  ];

  for (const metadata of invalidMetadata) {
    expect(() => session.setFromLogin(metadata as { sub: string; exp: bigint })).toThrow(
      /Session metadata/,
    );
    expect(session.claims()).toBeNull();
  }

  const claims = validClaims();
  session.setFromLogin(loginMetadata(claims));
  const before = stored();
  expect(() =>
    session.setFromLogin({
      sub: "replacement",
      exp: BigInt(Number.MAX_SAFE_INTEGER) + 1n,
    }),
  ).toThrow(/expiration/);
  expect(session.claims()).toEqual(claims);
  expect(stored()).toBe(before);
});

test("rejects invalid persisted, fractional, old, unavailable, and expired storage safely", () => {
  const now = Date.now();
  const invalidCachedValues = [
    { sub: "x", username: "y", exp: now / 1000 + 60 },
    { sub: "x", exp: now / 1000 + 0.5 },
    { sub: "x", exp: now / 1000 - 1 },
  ];
  for (const value of invalidCachedValues) {
    store(value);
    cleanup = session.initialize();
    expect(session.claims()).toBeNull();
    expect(stored()).toBeNull();
    cleanup();
    cleanup = undefined;
  }

  localStorage.setItem(SESSION_STORAGE_KEY, "not-json");
  cleanup = session.initialize();
  expect(session.claims()).toBeNull();
  expect(stored()).toBeNull();

  cleanup();
  breakStorage("getItem");
  expect(() => {
    cleanup = session.initialize();
  }).not.toThrow();
  expect(session.claims()).toBeNull();
  expect(() => session.setFromLogin(loginMetadata())).not.toThrow();
  expect(session.claims()).toBeTruthy();
});

test("restores cached claims before entering a render root and cleans up with it", () => {
  const claims = validClaims();
  store(claims);
  const cleanupSession = session.initialize();
  cleanup = cleanupSession;

  const dispose = createRoot((dispose) => {
    onCleanup(cleanupSession);
    expect(untrack(session.claims)).toEqual(claims);
    return dispose;
  });

  expect(vi.getTimerCount()).toBe(1);
  dispose();
  expect(vi.getTimerCount()).toBe(0);

  store(validClaims({ sub: "other-tab" }));
  window.dispatchEvent(new StorageEvent("storage", { key: SESSION_STORAGE_KEY }));
  expect(session.claims()).toEqual(claims);
});

test("storage write and removal failures do not prevent local login or logout", () => {
  cleanup = session.initialize();
  breakStorage("setItem");
  session.setFromLogin(loginMetadata());
  expect(session.claims()).toBeTruthy();
  expect(stored()).toBeNull();
  window.dispatchEvent(new Event("focus"));
  expect(session.claims()).toBeTruthy();
  document.dispatchEvent(new Event("visibilitychange", { bubbles: true }));
  expect(session.claims()).toBeTruthy();
  breakStorage("removeItem");
  expect(() => session.clear()).not.toThrow();
  expect(session.claims()).toBeNull();
});

test("increments revision for meaningful changes and allows same-claims login", () => {
  cleanup = session.initialize();
  const metadata = loginMetadata();
  const initial = session.revision();
  session.setFromLogin(metadata);
  const firstLogin = session.revision();
  session.setFromLogin(metadata);
  expect(firstLogin).toBe(initial + 1);
  expect(session.revision()).toBe(firstLogin + 1);
  session.clear();
  expect(session.revision()).toBe(firstLogin + 2);
  session.clear();
  expect(session.revision()).toBe(firstLogin + 2);
});

test("replaces the expiry timer when claims change", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 5 })));
  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 20 })));

  vi.advanceTimersByTime(6_000);
  expect(session.claims()).toBeTruthy();
  vi.advanceTimersByTime(14_000);
  expect(session.claims()).toBeNull();
});

test("does not let an old expiry timer clear a newer stored snapshot", () => {
  cleanup = session.initialize();
  const oldClaims = validClaims({ exp: nowInSeconds() + 5 });
  const newClaims = validClaims({ sub: "refreshed", exp: nowInSeconds() + 30 });
  session.setFromLogin(loginMetadata(oldClaims));

  store(newClaims);
  vi.advanceTimersByTime(6_000);
  expect(session.claims()).toEqual(newClaims);
});

test("tab resume adopts a newer cached session before expiring the old one", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 5 })));
  const renewed = validClaims({ sub: "renewed", exp: nowInSeconds() + 30 });
  store(renewed);
  sleep(6_000);
  window.dispatchEvent(new Event("focus"));
  expect(session.claims()).toEqual(renewed);
  expect(stored()).toBe(JSON.stringify(renewed));
});

test("rechecks expiration on focus, pageshow, and visible resume", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 10 })));
  sleep(11_000);
  window.dispatchEvent(new Event("focus"));
  expect(session.claims()).toBeNull();

  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 10 })));
  sleep(11_000);
  window.dispatchEvent(new Event("pageshow"));
  expect(session.claims()).toBeNull();

  session.setFromLogin(loginMetadata(validClaims({ exp: nowInSeconds() + 10 })));
  sleep(11_000);
  expect(document.visibilityState).toBe("visible");
  document.dispatchEvent(new Event("visibilitychange", { bubbles: true }));
  expect(session.claims()).toBeNull();
});

test("synchronizes cross-tab storage from the latest value, including key-null events", () => {
  cleanup = session.initialize();
  const first = validClaims({ sub: "first" });
  const second = validClaims({ sub: "second" });
  store(first);
  window.dispatchEvent(
    new StorageEvent("storage", { key: SESSION_STORAGE_KEY, newValue: JSON.stringify(second) }),
  );
  expect(session.claims()).toEqual(first);

  store(second);
  window.dispatchEvent(
    new StorageEvent("storage", { key: SESSION_STORAGE_KEY, newValue: JSON.stringify(first) }),
  );
  expect(session.claims()).toEqual(second);

  localStorage.removeItem(SESSION_STORAGE_KEY);
  window.dispatchEvent(new StorageEvent("storage", { key: null, newValue: null }));
  expect(session.claims()).toBeNull();
});

test("cleanup cancels timers and lifecycle listeners", () => {
  cleanup = session.initialize();
  const claims = validClaims({ exp: nowInSeconds() + 5 });
  session.setFromLogin(loginMetadata(claims));
  cleanup();
  cleanup = undefined;

  localStorage.removeItem(SESSION_STORAGE_KEY);
  window.dispatchEvent(new StorageEvent("storage", { key: SESSION_STORAGE_KEY }));
  expect(session.claims()).toEqual(claims);
  vi.advanceTimersByTime(10_000);
  expect(session.claims()).toEqual(claims);
});
