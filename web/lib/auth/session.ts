import { createSignal } from "solid-js";

export type SessionClaims = {
  sub: string;
  exp: number;
};

export const SESSION_STORAGE_KEY = "trevstack.session";

const MAX_TIMEOUT_DELAY = 2_147_483_647;
const CLAIM_KEYS = ["sub", "exp"] as const;

type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;
type BrowserWindow = Pick<Window, "addEventListener" | "removeEventListener"> & {
  localStorage?: StorageLike;
  document?: Pick<Document, "visibilityState">;
};

type StorageRead =
  | { kind: "unavailable" }
  | { kind: "missing" }
  | { kind: "invalid" }
  | { kind: "valid"; claims: SessionClaims };

// Solid 2 batches signal setters until an imperative boundary. Keep the
// authoritative values in plain memory so imperative callers (startup and RPC
// guards) observe a mutation synchronously, while the signal remains the
// reactive notification channel for components.
const [changeSignal, setChangeSignal] = createSignal(0);
let currentClaims: SessionClaims | null = null;
let currentRevision = 0;

let expiryTimer: ReturnType<typeof setTimeout> | undefined;
let initialized = false;
let activeCleanup: (() => void) | undefined;

function browserWindow(): BrowserWindow | null {
  const candidate = (globalThis as typeof globalThis & { window?: BrowserWindow }).window;
  return candidate && typeof candidate.addEventListener === "function" ? candidate : null;
}

function storage(): StorageLike | null {
  try {
    const browser = browserWindow();
    const candidate =
      browser?.localStorage ??
      (globalThis as typeof globalThis & { localStorage?: StorageLike }).localStorage;
    if (
      !candidate ||
      typeof candidate.getItem !== "function" ||
      typeof candidate.setItem !== "function" ||
      typeof candidate.removeItem !== "function"
    ) {
      return null;
    }
    return candidate;
  } catch {
    return null;
  }
}

function removePersistedValue() {
  try {
    storage()?.removeItem(SESSION_STORAGE_KEY);
  } catch {
    // Storage is optional. Keep the in-memory session when it is unavailable.
  }
}

function readPersistedValue(): StorageRead {
  const availableStorage = storage();
  if (!availableStorage) return { kind: "unavailable" };

  let value: string | null;
  try {
    value = availableStorage.getItem(SESSION_STORAGE_KEY);
  } catch {
    return { kind: "unavailable" };
  }

  if (value === null) return { kind: "missing" };

  try {
    const parsed: unknown = JSON.parse(value);
    const claims = validateClaims(parsed);
    if (!claims) {
      removePersistedValue();
      return { kind: "invalid" };
    }
    return { kind: "valid", claims };
  } catch {
    removePersistedValue();
    return { kind: "invalid" };
  }
}

function nowInSeconds() {
  return Date.now() / 1000;
}

function validateClaims(value: unknown): SessionClaims | null {
  if (typeof value !== "object" || value === null || Array.isArray(value)) return null;

  const record = value as Record<string, unknown>;
  const keys = Object.keys(record);
  if (keys.length !== CLAIM_KEYS.length || !CLAIM_KEYS.every((key) => keys.includes(key))) {
    return null;
  }

  if (
    typeof record.sub !== "string" ||
    record.sub.length === 0 ||
    typeof record.exp !== "number" ||
    !Number.isSafeInteger(record.exp) ||
    record.exp <= nowInSeconds()
  ) {
    return null;
  }

  return Object.freeze({
    sub: record.sub,
    exp: record.exp,
  });
}

function claimsFromLogin(metadata: { sub: string; exp: bigint }): SessionClaims {
  if (typeof metadata !== "object" || metadata === null || Array.isArray(metadata)) {
    throw new Error("Session metadata is invalid");
  }
  const value = metadata as unknown as Record<string, unknown>;
  if (typeof value.sub !== "string" || typeof value.exp !== "bigint") {
    throw new Error("Session metadata is invalid");
  }
  if (value.exp <= 0n || value.exp > BigInt(Number.MAX_SAFE_INTEGER)) {
    throw new Error("Session metadata expiration is invalid");
  }

  const claims = validateClaims({ sub: value.sub, exp: Number(value.exp) });
  if (!claims) throw new Error("Session metadata is invalid");
  return claims;
}

function sameClaims(left: SessionClaims | null, right: SessionClaims | null) {
  return left?.sub === right?.sub && left?.exp === right?.exp;
}

function cancelExpiryTimer() {
  if (expiryTimer !== undefined) {
    clearTimeout(expiryTimer);
    expiryTimer = undefined;
  }
}

function expireIfCurrent(expectedRevision: number) {
  if (expectedRevision !== currentRevision) return;
  const current = currentClaims;
  if (!current || current.exp > nowInSeconds()) {
    if (current) armExpiryTimer(current, expectedRevision);
    return;
  }

  // A storage event can be delayed. Adopt a newer snapshot before allowing an
  // old timer to clear a session that was refreshed in another tab.
  const persisted = readPersistedValue();
  if (persisted.kind === "valid" && !sameClaims(current, persisted.claims)) {
    setCurrentClaims(persisted.claims);
    return;
  }

  removePersistedValue();
  setCurrentClaims(null, false);
}

function armExpiryTimer(current: SessionClaims, expectedRevision = currentRevision) {
  cancelExpiryTimer();
  if (!initialized) return;

  const remaining = current.exp * 1000 - Date.now();
  if (!(remaining > 0)) {
    expireIfCurrent(expectedRevision);
    return;
  }

  expiryTimer = setTimeout(
    () => {
      expiryTimer = undefined;
      expireIfCurrent(expectedRevision);
    },
    Math.min(remaining, MAX_TIMEOUT_DELAY),
  );
}

function setCurrentClaims(next: SessionClaims | null, force = false) {
  const previous = currentClaims;
  if (!force && sameClaims(previous, next)) {
    if (next) armExpiryTimer(next);
    else cancelExpiryTimer();
    return;
  }

  cancelExpiryTimer();
  currentClaims = next;
  const nextRevision = currentRevision + 1;
  currentRevision = nextRevision;
  setChangeSignal(nextRevision);
  if (next) armExpiryTimer(next, nextRevision);
}

function syncFromStorage() {
  const result = readPersistedValue();
  if (result.kind === "unavailable") return;
  setCurrentClaims(result.kind === "valid" ? result.claims : null);
}

function checkExpiry() {
  expireIfCurrent(currentRevision);
}

function initialize(): () => void {
  activeCleanup?.();
  activeCleanup = undefined;
  initialized = true;

  syncFromStorage();
  checkExpiry();

  const browser = browserWindow();
  if (!browser) {
    const cleanup = () => {
      if (activeCleanup !== cleanup) return;
      activeCleanup = undefined;
      initialized = false;
      cancelExpiryTimer();
    };
    activeCleanup = cleanup;
    return cleanup;
  }

  const onStorage = (event: Event) => {
    const storageEvent = event as StorageEvent;
    if (storageEvent.key !== null && storageEvent.key !== SESSION_STORAGE_KEY) return;
    // Read the current value instead of trusting a potentially stale event payload.
    syncFromStorage();
  };
  const onResume = () => checkExpiry();
  const onVisibilityChange = () => {
    if (browser.document?.visibilityState === "visible") onResume();
  };

  browser.addEventListener("storage", onStorage);
  browser.addEventListener("focus", onResume);
  browser.addEventListener("pageshow", onResume);
  browser.addEventListener("visibilitychange", onVisibilityChange);

  const cleanup = () => {
    if (activeCleanup !== cleanup) return;
    activeCleanup = undefined;
    initialized = false;
    cancelExpiryTimer();
    browser.removeEventListener("storage", onStorage);
    browser.removeEventListener("focus", onResume);
    browser.removeEventListener("pageshow", onResume);
    browser.removeEventListener("visibilitychange", onVisibilityChange);
  };
  activeCleanup = cleanup;
  return cleanup;
}

export const session = {
  claims: () => {
    changeSignal();
    return currentClaims;
  },
  revision: () => {
    changeSignal();
    return currentRevision;
  },
  setFromLogin(metadata: { sub: string; exp: bigint }) {
    const claims = claimsFromLogin(metadata);
    try {
      storage()?.setItem(SESSION_STORAGE_KEY, JSON.stringify(claims));
    } catch {
      // Storage is optional. The in-memory session remains authoritative for this tab.
    }
    setCurrentClaims(claims, true);
  },
  clear() {
    removePersistedValue();
    setCurrentClaims(null);
  },
  initialize,
};
