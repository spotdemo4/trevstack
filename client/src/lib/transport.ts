import type { Interceptor } from '@connectrpc/connect';
import { createValidator } from '@bufbuild/protovalidate';
import { Code, ConnectError, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import { ItemService } from '$lib/connect/item/v1/item_pb';
import { AuthService } from '$lib/connect/user/v1/auth_pb';
import { UserService } from '$lib/connect/user/v1/user_pb';

const redirector: Interceptor = (next) => async (req) => {
	try {
		return await next(req);
	} catch (e) {
		const error = ConnectError.from(e);
		if (error.code === Code.Unauthenticated) {
			const redirectURL = new URL(page.url);
			redirectURL.pathname = '/auth';
			redirectURL.searchParams.append(
				'redir',
				encodeURIComponent(page.url.pathname + page.url.search)
			);

			await goto(redirectURL);
		}
		throw e;
	}
};

const transport = createConnectTransport({
	baseUrl: `${window.location.origin}/grpc`,
	interceptors: [redirector]
});

export const AuthClient = createClient(AuthService, transport);
export const UserClient = createClient(UserService, transport);
export const ItemClient = createClient(ItemService, transport);

export const Validator = createValidator();
