import { ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { GreetService } from "$connect/greet/v1/greet_pb";

const transport = createConnectTransport({
	baseUrl: `${window.location.origin}/grpc`,
});

type Result<T> = [T, null] | [null, ConnectError];

type SafeClient<T> = {
	[K in keyof T]: T[K] extends (...args: infer A) => Promise<infer R>
		? (...args: A) => Promise<Result<R>>
		: T[K];
};

function createSafeClient<T extends object>(client: T): SafeClient<T> {
	return new Proxy(client, {
		get(target, prop) {
			const value = target[prop as keyof T];
			if (typeof value === "function") {
				return async (...args: unknown[]) => {
					try {
						const result = await (
							value as (...a: unknown[]) => Promise<unknown>
						).apply(target, args);
						return [result, null];
					} catch (e) {
						return [null, ConnectError.from(e)];
					}
				};
			}
			return value;
		},
	}) as SafeClient<T>;
}

export const GreetClient = createSafeClient(
	createClient(GreetService, transport),
);
