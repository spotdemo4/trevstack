import type { DescService } from "@bufbuild/protobuf";
import { type Client, ConnectError, createClient, type Transport } from "@connectrpc/connect";

type Result<T> = [T, null] | [null, ConnectError];

type SafeClient<T> = {
  [K in keyof T]: T[K] extends (...args: infer A) => Promise<infer R>
    ? (...args: A) => Promise<Result<R>>
    : T[K];
};

export function createSafeClient<T extends DescService>(
  service: T,
  transport: Transport,
): SafeClient<Client<T>> {
  const client = createClient(service, transport);
  return new Proxy(client, {
    get(target, prop) {
      const value = target[prop as keyof Client<T>];
      if (typeof value === "function") {
        return async (...args: unknown[]) => {
          try {
            const result = await (value as (...a: unknown[]) => Promise<unknown>).apply(
              target,
              args,
            );
            return [result, null];
          } catch (e) {
            return [null, ConnectError.from(e)];
          }
        };
      }
      return value;
    },
  }) as SafeClient<Client<T>>;
}
