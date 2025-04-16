import type { User } from './connect/user/v1/user_pb';

export const userState: { user: User | undefined } = $state({
	user: undefined
});
