<script lang="ts">
	import { createApiReference } from '@scalar/api-reference';
	import { onDestroy, onMount } from 'svelte';

	let ref: ReturnType<typeof createApiReference>;
	let htmlref: HTMLDivElement;

	onMount(() => {
		ref = createApiReference(htmlref, {
			url: '/openapi/openapi.yaml',
			hideClientButton: true
		});
	});

	onDestroy(() => {
		ref.destroy();
	});
</script>

<svelte:head>
	<style>
		@import '@scalar/api-reference/style.css';
	</style>
</svelte:head>

<div bind:this={htmlref}></div>
