import { LoginRequestSchema } from "$connect/auth/v1/login_pb";
import { SignupRequestSchema } from "$connect/auth/v1/signup_pb";
import { getReturnPath, session } from "$lib/auth";
import { Card } from "$lib/card";
import { AuthClient } from "$lib/connect";
import { Form } from "$lib/form/form";
import { type MessageFormController, useForm } from "$lib/form/hook";
import { SubmitButton } from "$lib/form/submit-button";
import { TextField } from "$lib/form/text-field";
import { Tabs } from "$lib/tabs";
import { toaster } from "$lib/toast";
import { useNavigate } from "@solidjs/router";
import { Effect } from "effect";
import type { Component } from "solid-js";

export const Auth: Component = () => {
  const navigate = useNavigate();
  const returnTo = getReturnPath(window.location.search, window.location.origin);

  const loginForm = useForm(LoginRequestSchema, () => ({
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
              description: err.message,
            });
          },
        }),
        Effect.runPromise,
      ),
  }));

  const signupForm = useForm(SignupRequestSchema, () => ({
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

const LoginForm: Component<{
  form: MessageFormController<typeof LoginRequestSchema>;
}> = (props) => (
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

const SignupForm: Component<{
  form: MessageFormController<typeof SignupRequestSchema>;
}> = (props) => (
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

function rememberSession(response: { sub: string; exp: bigint }) {
  return Effect.try({
    try: () => session.setFromLogin({ sub: response.sub, exp: response.exp }),
    catch: () => new Error("Could not read your session. Please sign in again."),
  });
}
