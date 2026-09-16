import { createSignal } from "solid-js";

export type ToastType = "success" | "error" | "warning" | "info" | "loading";

export type ToastOptions = {
  title?: string;
  description?: string;
  duration?: number;
  closable?: boolean;
};

export type ToastRecord = ToastOptions & {
  id: number;
  type: ToastType;
};

type ToastState = "open" | "closed";

type DurationTimer = {
  handle?: ReturnType<typeof setTimeout>;
  remaining: number;
  startedAt?: number;
};

const MAX_TOASTS = 5;
const MAX_RETAINED_TOASTS = MAX_TOASTS * 2;
const EXIT_FALLBACK_MS = 200;

const [toastVersion, setToastVersion] = createSignal(0);
const [closingVersion, setClosingVersion] = createSignal(0);
const closingIds = new Set<number>();
const durationTimers = new Map<number, DurationTimer>();
const finalizationTimers = new Map<number, ReturnType<typeof setTimeout>>();
let records: ToastRecord[] = [];
let nextId = 1;
let paused = false;

const toasts = () => {
  toastVersion();
  return records;
};

const publishToasts = () => setToastVersion((current) => current + 1);
const publishClosing = () => setClosingVersion((current) => current + 1);

const clearDurationTimer = (id: number) => {
  const timer = durationTimers.get(id);
  if (timer?.handle) clearTimeout(timer.handle);
  durationTimers.delete(id);
};

const clearFinalizationTimer = (id: number) => {
  const timer = finalizationTimers.get(id);
  if (timer) clearTimeout(timer);
  finalizationTimers.delete(id);
};

const finalize = (id: number) => {
  clearDurationTimer(id);
  clearFinalizationTimer(id);

  const nextRecords = records.filter((toast) => toast.id !== id);
  if (nextRecords.length !== records.length) {
    records = nextRecords;
    if (records.length === 0) paused = false;
    publishToasts();
  }

  if (closingIds.delete(id)) publishClosing();
};

const dismiss = (id: number) => {
  if (!records.some((toast) => toast.id === id) || closingIds.has(id)) return;

  clearDurationTimer(id);
  closingIds.add(id);
  publishClosing();
  finalizationTimers.set(
    id,
    setTimeout(() => finalize(id), EXIT_FALLBACK_MS),
  );
};

const scheduleDurationTimer = (id: number) => {
  const timer = durationTimers.get(id);
  if (!timer || paused) return;

  timer.startedAt = Date.now();
  timer.handle = setTimeout(() => {
    durationTimers.delete(id);
    dismiss(id);
  }, timer.remaining);
};

const pauseAll = () => {
  if (paused) return;
  paused = true;
  const now = Date.now();

  for (const timer of durationTimers.values()) {
    if (!timer.handle || timer.startedAt === undefined) continue;
    clearTimeout(timer.handle);
    timer.remaining = Math.max(0, timer.remaining - (now - timer.startedAt));
    timer.handle = undefined;
    timer.startedAt = undefined;
  }
};

const resumeAll = () => {
  if (!paused) return;
  paused = false;

  const expiredIds: number[] = [];
  for (const [id, timer] of durationTimers) {
    if (timer.remaining <= 0) expiredIds.push(id);
    else scheduleDurationTimer(id);
  }

  for (const id of expiredIds) {
    durationTimers.delete(id);
    dismiss(id);
  }
};

const create = (type: ToastType, options: ToastOptions) => {
  const id = nextId++;
  const toast: ToastRecord = {
    id,
    type,
    duration: type === "loading" ? undefined : 5000,
    closable: true,
    ...options,
  };

  const openToasts = records.filter((current) => !closingIds.has(current.id));
  if (openToasts.length >= MAX_TOASTS) dismiss(openToasts[0].id);

  records = [...records, toast];
  publishToasts();

  const excess = records.length - MAX_RETAINED_TOASTS;
  if (excess > 0) {
    const staleIds = records
      .filter((current) => closingIds.has(current.id))
      .slice(0, excess)
      .map((current) => current.id);
    for (const staleId of staleIds) finalize(staleId);
  }

  if (toast.duration && toast.duration > 0) {
    durationTimers.set(id, { remaining: toast.duration });
    scheduleDurationTimer(id);
  }

  return id;
};

export const toaster = {
  toasts,
  dismiss,
  success: (options: ToastOptions) => create("success", options),
  error: (options: ToastOptions) => create("error", options),
  warning: (options: ToastOptions) => create("warning", options),
  info: (options: ToastOptions) => create("info", options),
  loading: (options: ToastOptions) => create("loading", options),
};

export const toastController = {
  state: (id: number): ToastState => {
    closingVersion();
    return closingIds.has(id) ? "closed" : "open";
  },
  finalize,
  pauseAll,
  resumeAll,
};
