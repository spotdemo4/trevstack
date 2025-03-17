import type { User } from "./services/user/v1/user_pb"

export let userState: { user: User | undefined } = $state({
	user: undefined
}); 