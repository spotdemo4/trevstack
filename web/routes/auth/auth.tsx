import { type LoginRequest, LoginRequestSchema } from "$connect/auth/v1/login_pb";
import { type SignupRequest, SignupRequestSchema } from "$connect/auth/v1/signup_pb";
import { Card } from "$lib/card";
import { AuthClient } from "$lib/connect";
import { Form } from "$lib/form/form";
import { type FormController, useForm } from "$lib/form/hook";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { toaster } from "$lib/toast";
import { create } from "@bufbuild/protobuf";
import { createStandardSchema } from "@bufbuild/protovalidate";
import { useNavigate } from "@solidjs/router";
import { Effect } from "effect";
import { Show, type Component, createSignal } from "solid-js";

export const Auth: Component = () => {
  const navigate = useNavigate();
  const [mode, setMode] = createSignal<"login" | "signup">("login");
  const returnTo = getReturnPath();

  const loginForm = useForm(() => ({
    defaultValues: { ...create(LoginRequestSchema) },
    validators: {
      onMount: createStandardSchema(LoginRequestSchema),
      onChange: createStandardSchema(LoginRequestSchema),
    },
    onSubmit: ({ value }) =>
      AuthClient.login({ username: value.username, password: value.password }).pipe(
        Effect.match({
          onSuccess: () => {
            toaster.success({ title: "Signed in" });
            navigate(returnTo, { replace: true });
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to sign in",
              description: err.message,
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  const signupForm = useForm(() => ({
    defaultValues: { ...create(SignupRequestSchema) },
    validators: {
      onMount: createStandardSchema(SignupRequestSchema),
      onChange: createStandardSchema(SignupRequestSchema),
    },
    onSubmit: ({ value }) =>
      Effect.gen(function* () {
        yield* AuthClient.signup({ username: value.username, password: value.password });
        yield* AuthClient.login({ username: value.username, password: value.password });
      }).pipe(
        Effect.match({
          onSuccess: () => {
            toaster.success({ title: "Account created" });
            navigate(returnTo, { replace: true });
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to create account",
              description: err.message,
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  return (
    <div class="flex h-body flex-col items-center justify-center gap-4 px-4">
      <h1 class="text-2xl font-bold">Welcome to TrevStack</h1>
      <Card class="w-full max-w-sm">
        <div class="mb-6 grid grid-cols-2 rounded-lg bg-ctp-crust p-1">
          <button
            type="button"
            aria-pressed={mode() === "login" ? "true" : "false"}
            class={tabClass(mode() === "login")}
            onClick={() => setMode("login")}
          >
            Login
          </button>
          <button
            type="button"
            aria-pressed={mode() === "signup" ? "true" : "false"}
            class={tabClass(mode() === "signup")}
            onClick={() => setMode("signup")}
          >
            Sign up
          </button>
        </div>

        <Show when={mode() === "login"} fallback={<SignupForm form={signupForm} />}>
          <LoginForm form={loginForm} />
        </Show>
      </Card>
    </div>
  );
};

const LoginForm: Component<{ form: FormController<LoginRequest> }> = (props) => (
  <Form form={props.form} class="flex-col">
    <TextField field={props.form.field("username")} label="Username" autocomplete="username" />
    <TextField
      field={props.form.field("password")}
      label="Password"
      type="password"
      autocomplete="current-password"
    />
    <SubmitButton form={props.form} label="Login" />
  </Form>
);

const SignupForm: Component<{ form: FormController<SignupRequest> }> = (props) => (
  <Form form={props.form} class="flex-col">
    <TextField field={props.form.field("username")} label="Username" autocomplete="username" />
    <TextField
      field={props.form.field("password")}
      label="Password"
      type="password"
      autocomplete="new-password"
    />
    <SubmitButton form={props.form} label="Create account" />
  </Form>
);

function getReturnPath() {
  const returnTo = new URLSearchParams(window.location.search).get("returnTo");
  if (!returnTo || !returnTo.startsWith("/")) {
    return "/";
  }

  try {
    const target = new URL(returnTo, window.location.origin);
    const normalizedPath = target.pathname.toLowerCase().replace(/\/+$/, "") || "/";
    if (target.origin !== window.location.origin || normalizedPath === "/auth") {
      return "/";
    }
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return "/";
  }
}

function tabClass(active: boolean) {
  return active
    ? "rounded-md bg-ctp-surface0 px-3 py-2 text-sm font-semibold text-ctp-text shadow-sm"
    : "rounded-md px-3 py-2 text-sm font-medium text-ctp-subtext0 hover:cursor-pointer hover:text-ctp-text";
}
