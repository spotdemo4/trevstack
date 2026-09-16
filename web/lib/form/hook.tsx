import { createSignal, onSettled } from "solid-js";

export type StandardSchemaIssue = {
  message: string;
  path?: readonly (PropertyKey | { key: PropertyKey })[];
};

type StandardSchemaResult<T> =
  | { value: T; issues?: undefined }
  | { issues: readonly StandardSchemaIssue[] };

export type StandardSchema<T> = {
  readonly "~standard": {
    readonly validate: (
      value: unknown,
    ) => StandardSchemaResult<T> | Promise<StandardSchemaResult<T>>;
  };
};

export type FieldController<T> = {
  readonly name: string;
  readonly value: () => T;
  readonly errors: () => string[];
  readonly isBlurred: () => boolean;
  readonly isValid: () => boolean;
  readonly invalid: () => boolean;
  readonly handleChange: (value: T) => void;
  readonly handleBlur: () => void;
};

export type FormController<T extends object> = {
  readonly values: () => T;
  readonly errors: () => string[];
  readonly isSubmitting: () => boolean;
  readonly isValidating: () => boolean;
  readonly isValid: () => boolean;
  readonly isDefaultValue: () => boolean;
  readonly canSubmit: () => boolean;
  readonly field: <K extends keyof T & string>(name: K) => FieldController<T[K]>;
  readonly reset: () => void;
  readonly handleSubmit: () => Promise<boolean>;
};

type FormOptions<T extends object> = {
  defaultValues: T;
  validators?: {
    onMount?: StandardSchema<T>;
    onChange?: StandardSchema<T>;
  };
  onSubmit: (args: { value: T }) => void | Promise<unknown>;
};

type FormOptionsAccessor<T extends object> = FormOptions<T> | (() => FormOptions<T>);

type ErrorState = Record<string, string[]>;

function issueKey(issue: StandardSchemaIssue): string | undefined {
  const first = issue.path?.[0];
  if (typeof first === "string" || typeof first === "number") return String(first);
  if (first && typeof first === "object" && "key" in first) return String(first.key);
  return undefined;
}

function errorsFromIssues(issues: readonly StandardSchemaIssue[]): ErrorState {
  const errors: ErrorState = {};
  for (const issue of issues) {
    const key = issueKey(issue) ?? "";
    (errors[key] ??= []).push(issue.message);
  }
  return errors;
}

function hasErrors(errors: ErrorState): boolean {
  return Object.values(errors).some((messages) => messages.length > 0);
}

function sameValues<T extends object>(a: T, b: T): boolean {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
  return [...keys].every((key) => Object.is(a[key as keyof T], b[key as keyof T]));
}

export function useForm<T extends object>(options: FormOptionsAccessor<T>): FormController<T> {
  const config = typeof options === "function" ? options() : options;
  const initialValues = { ...config.defaultValues } as Exclude<T, Function>;
  const [values, setValues] = createSignal<T>(initialValues);
  const [errorState, setErrorState] = createSignal<ErrorState>({});
  const [blurred, setBlurred] = createSignal<Record<string, boolean>>({});
  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [isValidating, setIsValidating] = createSignal(false);
  let validationId = 0;

  const validate = async (schema: StandardSchema<T> | undefined, value: T): Promise<boolean> => {
    if (!schema) {
      setErrorState({});
      return true;
    }

    const currentValidation = ++validationId;
    setIsValidating(true);
    try {
      const result = await schema["~standard"].validate(value);
      if (currentValidation !== validationId) return !hasErrors(errorState());
      const errors = "issues" in result && result.issues ? errorsFromIssues(result.issues) : {};
      setErrorState(errors);
      return !hasErrors(errors);
    } finally {
      if (currentValidation === validationId) setIsValidating(false);
    }
  };

  const validateOnChange = (value: T): void => {
    const schema = config.validators?.onChange;
    if (schema) void validate(schema, value);
  };

  const field = <K extends keyof T & string>(name: K): FieldController<T[K]> => ({
    name,
    value: () => values()[name],
    errors: () => errorState()[name] ?? [],
    isBlurred: () => blurred()[name] === true,
    isValid: () => (errorState()[name] ?? []).length === 0,
    invalid: () => blurred()[name] === true && (errorState()[name] ?? []).length > 0,
    handleChange: (value) => {
      const next = { ...values(), [name]: value } as T;
      setValues(() => next);
      validateOnChange(next);
    },
    handleBlur: () => setBlurred((current) => ({ ...current, [name]: true })),
  });

  const reset = (): void => {
    validationId += 1;
    const next = { ...config.defaultValues } as T;
    setValues(() => next);
    setErrorState({});
    setBlurred({});
    setIsValidating(false);
    if (config.validators?.onMount) void validate(config.validators.onMount, next);
  };

  const handleSubmit = async (): Promise<boolean> => {
    setBlurred((current) => {
      const next = { ...current };
      for (const key of Object.keys(values())) next[key] = true;
      return next;
    });

    const schema = config.validators?.onChange ?? config.validators?.onMount;
    if (!(await validate(schema, values()))) return false;

    setIsSubmitting(true);
    try {
      await config.onSubmit({ value: values() });
      return true;
    } finally {
      setIsSubmitting(false);
    }
  };

  onSettled(() => {
    if (config.validators?.onMount) void validate(config.validators.onMount, values());
  });

  return {
    values,
    errors: () => errorState()[""] ?? [],
    isSubmitting,
    isValidating,
    isValid: () => !hasErrors(errorState()),
    isDefaultValue: () => sameValues(values(), config.defaultValues),
    canSubmit: () => !isSubmitting() && !isValidating() && !hasErrors(errorState()),
    field,
    reset,
    handleSubmit,
  };
}
