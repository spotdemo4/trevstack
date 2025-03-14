import { createConnectTransport } from "@connectrpc/connect-web"
import { Code, ConnectError, createClient, type Interceptor } from "@connectrpc/connect"
import { AuthService } from "$lib/services/user/v1/auth_pb";
import { UserService } from "$lib/services/user/v1/user_pb";
import { ItemService } from "$lib/services/item/v1/item_pb";
import { goto } from "$app/navigation";

const redirector: Interceptor = (next) => async (req) => {
    try {
        return await next(req);
    } catch (e) {
        const error = ConnectError.from(e);
        if (error.code === Code.Unauthenticated) {
            await goto('/auth');
        }
        throw e;
    }
};

const transport = createConnectTransport({
    baseUrl: `${window.location.origin}/grpc`,
    interceptors: [redirector],
});

export const AuthClient = createClient(AuthService, transport);
export const UserClient = createClient(UserService, transport);
export const ItemClient = createClient(ItemService, transport);