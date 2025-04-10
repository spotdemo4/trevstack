<script lang="ts">
	import { userState } from '$lib/sharedState.svelte';
	import { Avatar } from 'bits-ui';
</script>

<Avatar.Root class="flex h-full w-full items-center justify-center">
	<Avatar.Image
		src={userState.user?.profilePictureId ? "/file/" + userState.user.profilePictureId : null}
		alt={`${userState.user?.username}'s avatar`}
		class="rounded-full"
	/>
	<Avatar.Fallback class="font-medium uppercase"
		>{userState.user?.username.substring(0, 2)}</Avatar.Fallback
	>
</Avatar.Root>
