import { strict as assert } from "node:assert";
import { afterEach, beforeEach, test } from "node:test";

import { createRoot, onCleanup, untrack } from "solid-js";

import { SESSION_STORAGE_KEY, session, type SessionClaims } from "./session.ts";

const NOW = 1_700_000_000_000;
const realNow = Date.now;
const realSetTimeout = globalThis.setTimeout;
const realClearTimeout = globalThis.clearTimeout;

type Listener = (event: Event) => void;

class FakeStorage {
  readonly values = new Map<string, string>();
  throwOnGet = false;
  throwOnSet = false;
  throwOnRemove = false;

  getItem(key: string) {
    if (this.throwOnGet) throw new Error("storage unavailable");
    return this.values.get(key) ?? null;
  }

  setItem(key: string, value: string) {
    if (this.throwOnSet) throw new Error("storage unavailable");
    this.values.set(key, value);
  }

  removeItem(key: string) {
    if (this.throwOnRemove) throw new Error("storage unavailable");
    this.values.delete(key);
  }
}

class FakeWindow {
  readonly listeners = new Map<string, Set<Listener>>();
  readonly document = { visibilityState: "visible" as Document["visibilityState"] };
  readonly localStorage: FakeStorage;

  constructor(localStorage: FakeStorage) {
    this.localStorage = localStorage;
  }

  addEventListener(type: string, listener: Listener) {
    let listeners = this.listeners.get(type);
    if (!listeners) {
      listeners = new Set();
      this.listeners.set(type, listeners);
    }
    listeners.add(listener);
  }

  removeEventListener(type: string, listener: Listener) {
    this.listeners.get(type)?.delete(listener);
  }

  emit(type: string, event: Partial<StorageEvent> = {}) {
    const value = { type, ...event } as StorageEvent;
    for (const listener of this.listeners.get(type) ?? []) listener(value);
  }

  listenerCount(type: string) {
    return this.listeners.get(type)?.size ?? 0;
  }
}

let clock = NOW;
let nextTimerId = 1;
const timers = new Map<number, { at: number; callback: () => void }>();
let browser: FakeWindow;
let storage: FakeStorage;
let cleanup: (() => void) | undefined;

function setFakeTimeout(callback: () => void, delay: number) {
  const id = nextTimerId++;
  timers.set(id, { at: clock + delay, callback });
  return id;
}

function clearFakeTimeout(id: number | undefined) {
  if (id !== undefined) timers.delete(id);
}

function advance(ms: number) {
  const target = clock + ms;
  while (true) {
    const due = [...timers.entries()]
      .filter(([, timer]) => timer.at <= target)
      .sort(([, left], [, right]) => left.at - right.at)[0];
    if (!due) break;
    const [id, timer] = due;
    timers.delete(id);
    clock = timer.at;
    timer.callback();
  }
  clock = target;
}

function validClaims(overrides: Partial<SessionClaims> = {}): SessionClaims {
  return {
    sub: "user-1",
    exp: Math.floor(clock / 1000) + 60,
    ...overrides,
  };
}

function loginMetadata(claims: SessionClaims = validClaims()) {
  return { sub: claims.sub, exp: BigInt(claims.exp) };
}

beforeEach(() => {
  clock = NOW;
  nextTimerId = 1;
  timers.clear();
  storage = new FakeStorage();
  browser = new FakeWindow(storage);
  const runtime = globalThis as unknown as {
    window?: FakeWindow;
    setTimeout: typeof setTimeout;
    clearTimeout: typeof clearTimeout;
  };
  runtime.window = browser;
  (Date as typeof Date).now = () => clock;
  runtime.setTimeout = setFakeTimeout as unknown as typeof setTimeout;
  runtime.clearTimeout = clearFakeTimeout as unknown as typeof clearTimeout;
});

afterEach(() => {
  cleanup?.();
  cleanup = undefined;
  session.clear();
  timers.clear();
  const runtime = globalThis as unknown as { window?: FakeWindow };
  delete runtime.window;
  Date.now = realNow;
  globalThis.setTimeout = realSetTimeout;
  globalThis.clearTimeout = realClearTimeout;
});

void test("accepts explicit login metadata, persists only the exact allowlist, and ignores extras", () => {
  const claims = validClaims({ sub: "用户/é" });
  cleanup = session.initialize();
  const response = {
    ...loginMetadata(claims),
    jwt: "ignored.jwt.value",
    arbitrary: { ignored: true },
  };
  session.setFromLogin(response);

  assert.deepEqual(session.claims(), claims);
  assert.deepEqual(JSON.parse(storage.values.get(SESSION_STORAGE_KEY)!), claims);
  assert.equal(storage.values.get(SESSION_STORAGE_KEY), JSON.stringify(claims));
  assert.deepEqual(Object.keys(JSON.parse(storage.values.get(SESSION_STORAGE_KEY)!)), [
    "sub",
    "exp",
  ]);
});

void test("rejects invalid login metadata before changing the session", () => {
  cleanup = session.initialize();
  const invalidMetadata: unknown[] = [
    null,
    undefined,
    [],
    { exp: BigInt(clock / 1000 + 60) },
    { sub: "user-1" },
    { sub: "user-1", exp: "future" },
    { sub: "user-1", exp: clock / 1000 + 60 },
    { sub: "user-1", exp: BigInt(clock / 1000) },
    { sub: "user-1", exp: BigInt(clock / 1000 - 1) },
    { sub: "", exp: BigInt(Math.floor(clock / 1000) + 1) },
    { sub: "user-1", exp: 0n },
    { sub: "user-1", exp: -1n },
    { sub: "user-1", exp: BigInt(Number.MAX_SAFE_INTEGER) + 1n },
  ];

  for (const metadata of invalidMetadata) {
    assert.throws(
      () => session.setFromLogin(metadata as { sub: string; exp: bigint }),
      /Session metadata/,
    );
    assert.equal(session.claims(), null);
  }

  const claims = validClaims();
  session.setFromLogin(loginMetadata(claims));
  const stored = storage.values.get(SESSION_STORAGE_KEY);
  assert.throws(
    () =>
      session.setFromLogin({
        sub: "replacement",
        exp: BigInt(Number.MAX_SAFE_INTEGER) + 1n,
      }),
    /expiration/,
  );
  assert.deepEqual(session.claims(), claims);
  assert.equal(storage.values.get(SESSION_STORAGE_KEY), stored);
});

void test("rejects invalid persisted, fractional, old, unavailable, and expired storage safely", () => {
  const invalidCachedValues = [
    { sub: "x", username: "y", exp: clock / 1000 + 60 },
    { sub: "x", exp: clock / 1000 + 0.5 },
    { sub: "x", exp: clock / 1000 - 1 },
  ];
  for (const value of invalidCachedValues) {
    storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(value));
    cleanup = session.initialize();
    assert.equal(session.claims(), null);
    assert.equal(storage.values.has(SESSION_STORAGE_KEY), false);
    cleanup();
    cleanup = undefined;
  }

  storage.values.set(SESSION_STORAGE_KEY, "not-json");
  cleanup = session.initialize();
  assert.equal(session.claims(), null);
  assert.equal(storage.values.has(SESSION_STORAGE_KEY), false);

  cleanup();
  storage.throwOnGet = true;
  assert.doesNotThrow(() => {
    cleanup = session.initialize();
  });
  assert.equal(session.claims(), null);
  assert.doesNotThrow(() => session.setFromLogin(loginMetadata()));
  assert.ok(session.claims());
});

void test("restores cached claims before entering a render root and cleans up with it", () => {
  const claims = validClaims();
  storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(claims));
  const cleanupSession = session.initialize();
  cleanup = cleanupSession;

  const dispose = createRoot((dispose) => {
    onCleanup(cleanupSession);
    assert.deepEqual(untrack(session.claims), claims);
    return dispose;
  });

  assert.equal(browser.listenerCount("storage"), 1);
  assert.equal(timers.size, 1);
  dispose();
  assert.equal(browser.listenerCount("storage"), 0);
  assert.equal(timers.size, 0);
});

void test("storage write and removal failures do not prevent local login or logout", () => {
  cleanup = session.initialize();
  storage.throwOnSet = true;
  session.setFromLogin(loginMetadata());
  assert.ok(session.claims());
  assert.equal(storage.values.has(SESSION_STORAGE_KEY), false);
  browser.emit("focus");
  assert.ok(session.claims());
  browser.emit("visibilitychange");
  assert.ok(session.claims());
  storage.throwOnRemove = true;
  assert.doesNotThrow(() => session.clear());
  assert.equal(session.claims(), null);
});

void test("increments revision for meaningful changes and allows same-claims login", () => {
  cleanup = session.initialize();
  const metadata = loginMetadata();
  const initial = session.revision();
  session.setFromLogin(metadata);
  const firstLogin = session.revision();
  session.setFromLogin(metadata);
  assert.equal(firstLogin, initial + 1);
  assert.equal(session.revision(), firstLogin + 1);
  session.clear();
  assert.equal(session.revision(), firstLogin + 2);
  session.clear();
  assert.equal(session.revision(), firstLogin + 2);
});

void test("replaces the expiry timer when claims change", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 5 })));
  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 20 })));

  advance(6_000);
  assert.ok(session.claims());
  advance(14_000);
  assert.equal(session.claims(), null);
});

void test("does not let an old expiry timer clear a newer stored snapshot", () => {
  cleanup = session.initialize();
  const oldClaims = validClaims({ exp: Math.floor(clock / 1000) + 5 });
  const newClaims = validClaims({ sub: "refreshed", exp: Math.floor(clock / 1000) + 30 });
  session.setFromLogin(loginMetadata(oldClaims));

  storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(newClaims));
  advance(6_000);
  assert.deepEqual(session.claims(), newClaims);
});

void test("tab resume adopts a newer cached session before expiring the old one", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 5 })));
  const renewed = validClaims({ sub: "renewed", exp: Math.floor(clock / 1000) + 30 });
  storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(renewed));
  clock += 6_000;
  browser.emit("focus");
  assert.deepEqual(session.claims(), renewed);
  assert.equal(storage.values.get(SESSION_STORAGE_KEY), JSON.stringify(renewed));
});

void test("rechecks expiration on focus, pageshow, and visible resume", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 10 })));
  clock += 11_000;
  browser.emit("focus");
  assert.equal(session.claims(), null);

  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 10 })));
  clock += 11_000;
  browser.emit("pageshow");
  assert.equal(session.claims(), null);

  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 10 })));
  clock += 11_000;
  browser.document.visibilityState = "visible";
  browser.emit("visibilitychange");
  assert.equal(session.claims(), null);
});

void test("synchronizes cross-tab storage from the latest value, including key-null events", () => {
  cleanup = session.initialize();
  const first = validClaims({ sub: "first" });
  const second = validClaims({ sub: "second" });
  storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(first));
  browser.emit("storage", { key: SESSION_STORAGE_KEY, newValue: JSON.stringify(second) });
  assert.deepEqual(session.claims(), first);

  storage.values.set(SESSION_STORAGE_KEY, JSON.stringify(second));
  browser.emit("storage", { key: SESSION_STORAGE_KEY, newValue: JSON.stringify(first) });
  assert.deepEqual(session.claims(), second);

  storage.values.delete(SESSION_STORAGE_KEY);
  browser.emit("storage", { key: null, newValue: null });
  assert.equal(session.claims(), null);
});

void test("cleanup cancels timers and lifecycle listeners", () => {
  cleanup = session.initialize();
  session.setFromLogin(loginMetadata(validClaims({ exp: Math.floor(clock / 1000) + 5 })));
  assert.equal(browser.listenerCount("storage"), 1);
  cleanup();
  cleanup = undefined;
  assert.equal(browser.listenerCount("storage"), 0);
  advance(10_000);
  assert.ok(session.claims());
});
