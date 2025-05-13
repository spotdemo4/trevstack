import type { DescMessage, DescService, MessageInitShape, MessageShape } from '@bufbuild/protobuf';
import type { Violation } from '@bufbuild/protovalidate';
import type { Client } from '@connectrpc/connect';
import type { Action } from 'svelte/action';
import { create } from '@bufbuild/protobuf';
import { ValidationError } from '@bufbuild/protovalidate';
import { ConnectError } from '@connectrpc/connect';
import { Validator } from '../transport';

type Options<Input extends DescMessage, Output extends DescMessage> = {
	init?: MessageInitShape<Input>;
	start?: boolean;
	reset?: boolean;
	onSubmit?: (formData: FormData, input: MessageShape<Input>) => Promise<MessageShape<Input>>;
	onResult?: (result: MessageShape<Output>) => void;
	onError?: (err: Violation[] | ConnectError) => void;
};

type Violations<Field> = {
	[field in keyof Field]?: Violation[];
};

export function coolForm<Service extends DescService, Method extends Service['methods'][number]>(
	client: Client<Service>,
	method: Method,
	options?: Options<Method['input'], Method['output']>
) {
	const input = $state(create(method.input as Method['input'], options?.init));
	const output = $state(create(method.output as Method['output']));
	const errors: Violations<Method['input']['field']> & { form?: ConnectError } = $state({});
	let loading = $state(false);

	const validate = () => {
		// Delete existing errors
		for (const key in errors) {
			delete errors[key];
		}

		try {
			Validator.validate(method['input'], input);
		} catch (e) {
			if (!(e instanceof ValidationError)) {
				throw e;
			}

			// Map violation errors to errors rune
			for (const violation of e.violations) {
				for (const field of violation.field) {
					if (!('localName' in field)) {
						continue;
					}

					// Create localName property if it doesn't exist
					if (!errors[field.localName]) {
						Object.assign(errors, {
							[field.localName]: [violation]
						});
					} else {
						errors[field.localName]?.push(violation);
					}
				}
			}

			return e.violations;
		}

		return [];
	};

	// When a request is successful
	const success = (response: MessageShape<Method['output']>) => {
		loading = false;

		// Send the response up
		options?.onResult?.(response);

		// Set the response
		Object.assign(output, response);

		// If we want to reset the input
		if (options && (options.reset == undefined || options.reset)) {
			const cleared = create(method.input as Method['input'], options?.init);
			Object.assign(input, cleared);
		}
	};

	// When a request fails
	const fail = (err: Violation[] | ConnectError | Error) => {
		loading = false;

		// It's a Violation[]
		if (Array.isArray(err)) {
			// Send the error up
			options?.onError?.(err);
			return;
		}

		// It's a ConnectError
		if (err instanceof ConnectError) {
			// Assign it to the form
			errors.form = err;

			// Send the error up
			options?.onError?.(err);
			return;
		}

		throw err;
	};

	const submit = () => {
		loading = true;

		// Validate
		const validationErrors = validate();
		if (validationErrors.length > 0) {
			fail(validationErrors);
			return;
		}

		// Send response
		if (method.methodKind == 'unary') {
			// @ts-expect-error I can't figure out how to make this typescript compliant
			const response = client[method.localName]($state.snapshot(input)) as Promise<
				MessageShape<Method['output']>
			>;

			response
				.then((resp) => {
					success(resp);
				})
				.catch((err) => {
					fail(err);
				});
		}
	};

	// A nice action to give to forms to run submit() on submit
	const impair: Action<HTMLFormElement> = (form) => {
		$effect(() => {
			form.onsubmit = (event) => {
				event.preventDefault();

				if (options?.onSubmit) {
					const formData = new FormData(form);
					options.onSubmit(formData, input).then((i) => {
						Object.assign(input, i);
						submit();
					});

					return;
				}

				submit();
			};

			return () => {
				form.onsubmit = () => {};
			};
		});
	};

	if (options?.start) {
		submit();
	}

	return {
		input,
		output,
		errors,
		loading: () => loading,
		submit,
		validate,
		impair
	};
}
