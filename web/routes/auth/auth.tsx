import type { LoginResponse } from "$gen/auth/v1/login_pb";
import { getReturnPath, session } from "$lib/auth";
import { Card } from "$lib/card";
import { errorDescription } from "$lib/effect";
import { Form } from "$lib/form/form";
import { type FormController, useForm } from "$lib/form/hook";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { AuthClient } from "$lib/rpc";
import { Tabs } from "$lib/tabs";
import { toaster } from "$lib/toast";
import { useNavigate } from "@solidjs/router";
import { Effect, Schema } from "effect";
import type { Component } from "solid-js";

const utf8 = new TextEncoder();

// The same rules as the server's, so most mistakes are caught before a request.
const Credentials = Schema.Struct({
  username: Schema.String.check(
    Schema.isMinLength(3, { message: "Username must be at least 3 characters" }),
    Schema.isMaxLength(64, { message: "Username must be at most 64 characters" }),
    Schema.isPattern(/^[A-Za-z0-9][A-Za-z0-9_-]*$/, {
      message:
        "Username must start with a letter or digit and contain only letters, digits, underscores, and hyphens",
    }),
  ).annotateKey({ messageMissingKey: "Username is required" }),
  password: Schema.String.check(
    Schema.isMinLength(8, { message: "Password must be at least 8 characters" }),
    Schema.makeFilter(
      (password) => utf8.encode(password).length <= 72 || "Password must be at most 72 bytes",
    ),
  ).annotateKey({ messageMissingKey: "Password is required" }),
});

const credentialsSchema = Schema.toStandardSchemaV1(Credentials);

type CredentialsForm = FormController<Partial<typeof Credentials.Encoded>>;

export const Auth: Component = () => {
  const navigate = useNavigate();
  const returnTo = getReturnPath(window.location.search, window.location.origin);

  const loginForm = useForm(credentialsSchema, () => ({
    validation: {
      onMount: true,
      onChange: true,
    },
    onSubmit: ({ value }) =>
      AuthClient.login({ username: value.username, password: value.password }).pipe(
        Effect.tap(rememberSession),
        Effect.match({
          onSuccess: () => {
            toaster.success({ title: "Signed in" });
            navigate(returnTo, { replace: true });
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to sign in",
              description: errorDescription(err),
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  const signupForm = useForm(credentialsSchema, () => ({
    validation: {
      onMount: true,
      onChange: true,
    },
    onSubmit: ({ value }) =>
      Effect.gen(function* () {
        yield* AuthClient.signup({ username: value.username, password: value.password });
        return yield* AuthClient.login({ username: value.username, password: value.password });
      }).pipe(
        Effect.tap(rememberSession),
        Effect.match({
          onSuccess: () => {
            toaster.success({ title: "Account created" });
            navigate(returnTo, { replace: true });
          },
          onFailure: (err) => {
            toaster.error({
              title: "Failed to create account",
              description: errorDescription(err),
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
        <Tabs.Root defaultValue="login">
          <Tabs.List aria-label="Authentication" class="grid w-full grid-cols-2">
            <Tabs.Trigger value="login">Login</Tabs.Trigger>
            <Tabs.Trigger value="signup">Sign up</Tabs.Trigger>
          </Tabs.List>
          <Tabs.Content value="login">
            <LoginForm form={loginForm} />
          </Tabs.Content>
          <Tabs.Content value="signup">
            <SignupForm form={signupForm} />
          </Tabs.Content>
        </Tabs.Root>
      </Card>
    </div>
  );
};

const LoginForm: Component<{ form: CredentialsForm }> = (props) => (
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

const SignupForm: Component<{ form: CredentialsForm }> = (props) => (
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

function rememberSession(response: LoginResponse) {
  return Effect.try({
    try: () => session.setFromLogin(response),
    catch: () => new Error("Could not read your session. Please sign in again."),
  });
}
