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

const [toasts, setToasts] = createSignal<ToastRecord[]>([]);
const timers = new Map<number, ReturnType<typeof setTimeout>>();
let nextId = 1;

const dismiss = (id: number) => {
  const timer = timers.get(id);
  if (timer) clearTimeout(timer);
  timers.delete(id);
  setToasts((current) => current.filter((toast) => toast.id !== id));
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

  setToasts((current) => [...current.slice(-4), toast]);
  if (toast.duration && toast.duration > 0) {
    timers.set(
      id,
      setTimeout(() => dismiss(id), toast.duration),
    );
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
