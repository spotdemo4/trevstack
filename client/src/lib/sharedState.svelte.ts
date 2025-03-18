import type { User } from './services/user/v1/user_pb';

export const userState: { user: User | undefined } = $state({
	user: undefined
});
