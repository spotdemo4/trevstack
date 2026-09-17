import type { DescMessage, MessageInitShape, MessageValidType } from "@bufbuild/protobuf";
import { createStandardSchemaInit } from "@bufbuild/protovalidate";
import { createSignal, onSettled } from "solid-js";

export type StandardSchemaIssue = {
  message: string;
  path?: readonly (PropertyKey | { key: PropertyKey })[];
};

type StandardSchemaResult<T> =
  | { value: T; issues?: undefined }
  | { issues: readonly StandardSchemaIssue[] };

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

export type MessageFormController<Desc extends DescMessage> = FormController<
  MessageInitShape<Desc>
>;

type ValidationTiming = {
  readonly onMount?: boolean;
  readonly onChange?: boolean;
};

type FormOptions<T extends object> = {
  defaultValues: T;
  onSubmit: (args: { value: T }) => void | Promise<unknown>;
};

type MessageFormOptions<Desc extends DescMessage> = {
  defaultValues?: MessageInitShape<Desc>;
  validation?: ValidationTiming;
  onSubmit: (args: { value: MessageValidType<Desc> }) => void | Promise<unknown>;
};

type OptionsAccessor<T extends object> = T | (() => T);

type InternalFormOptions<TInput extends object, TOutput> = {
  defaultValues: TInput;
  validation: Required<ValidationTiming> & { enabled: boolean };
  validate: (
    value: TInput,
  ) => StandardSchemaResult<TOutput> | Promise<StandardSchemaResult<TOutput>>;
  onSubmit: (args: { value: TOutput }) => void | Promise<unknown>;
};

type ErrorState = Record<string, string[]>;

type ValidationResult<T> = { valid: true; value: T } | { valid: false };

function resolveOptions<T extends object>(options: OptionsAccessor<T>): T {
  return typeof options === "function" ? options() : options;
}

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

export function useForm<Desc extends DescMessage>(
  schema: Desc,
  options: OptionsAccessor<MessageFormOptions<Desc>>,
): MessageFormController<Desc>;
export function useForm<T extends object>(
  options: OptionsAccessor<FormOptions<T>>,
): FormController<T>;
export function useForm<T extends object, Desc extends DescMessage>(
  ...args:
    | [options: OptionsAccessor<FormOptions<T>>]
    | [schema: Desc, options: OptionsAccessor<MessageFormOptions<Desc>>]
): FormController<T> | MessageFormController<Desc> {
  if (args.length === 1) {
    const config = resolveOptions(args[0]);
    return createForm({
      defaultValues: config.defaultValues,
      validation: { enabled: false, onMount: false, onChange: false },
      validate: (value) => ({ value }),
      onSubmit: config.onSubmit,
    });
  }

  const [schema, options] = args;
  const config = resolveOptions(options);
  const validator = createStandardSchemaInit(schema);
  const defaultValues = (config.defaultValues ?? {}) as MessageInitShape<Desc>;

  return createForm({
    defaultValues,
    validation: {
      enabled: true,
      onMount: config.validation?.onMount ?? false,
      onChange: config.validation?.onChange ?? false,
    },
    validate: (value) => validator["~standard"].validate(value),
    onSubmit: config.onSubmit,
  });
}

function createForm<TInput extends object, TOutput>(
  config: InternalFormOptions<TInput, TOutput>,
): FormController<TInput> {
  const initialValues = { ...config.defaultValues } as Exclude<TInput, Function>;
  const [values, setValues] = createSignal<TInput>(initialValues);
  const [errorState, setErrorState] = createSignal<ErrorState>({});
  const [blurred, setBlurred] = createSignal<Record<string, boolean>>({});
  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [isValidating, setIsValidating] = createSignal(false);
  let validationId = 0;

  const validate = async (value: TInput): Promise<ValidationResult<TOutput>> => {
    if (!config.validation.enabled) {
      setErrorState({});
      const result = await config.validate(value);
      return "value" in result ? { valid: true, value: result.value } : { valid: false };
    }

    const currentValidation = ++validationId;
    setIsValidating(true);
    try {
      const result = await config.validate(value);
      if (currentValidation !== validationId) return { valid: false };
      const errors = "issues" in result && result.issues ? errorsFromIssues(result.issues) : {};
      setErrorState(errors);
      if (hasErrors(errors) || !("value" in result)) return { valid: false };
      return { valid: true, value: result.value };
    } finally {
      if (currentValidation === validationId) setIsValidating(false);
    }
  };

  const validateOnChange = (value: TInput): void => {
    if (config.validation.onChange) void validate(value);
  };

  const field = <K extends keyof TInput & string>(name: K): FieldController<TInput[K]> => ({
    name,
    value: () => values()[name],
    errors: () => errorState()[name] ?? [],
    isBlurred: () => blurred()[name] === true,
    isValid: () => (errorState()[name] ?? []).length === 0,
    invalid: () => blurred()[name] === true && (errorState()[name] ?? []).length > 0,
    handleChange: (value) => {
      const next = { ...values(), [name]: value } as TInput;
      setValues(() => next);
      validateOnChange(next);
    },
    handleBlur: () => setBlurred((current) => ({ ...current, [name]: true })),
  });

  const reset = (): void => {
    validationId += 1;
    const next = { ...config.defaultValues } as TInput;
    setValues(() => next);
    setErrorState({});
    setBlurred({});
    setIsValidating(false);
    if (config.validation.onMount) void validate(next);
  };

  const handleSubmit = async (): Promise<boolean> => {
    const value = values();
    setBlurred((current) => {
      const next = { ...current };
      for (const key of Object.keys(value)) next[key] = true;
      return next;
    });

    const result = await validate(value);
    if (!result.valid) return false;

    setIsSubmitting(true);
    try {
      await config.onSubmit({ value: result.value });
      return true;
    } finally {
      setIsSubmitting(false);
    }
  };

  onSettled(() => {
    if (config.validation.onMount) void validate(values());
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
