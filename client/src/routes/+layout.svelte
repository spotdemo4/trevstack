<script>
	import '../app.css';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();
</script>

<Toaster toastOptions={{
	classes: {
		toast: '!bg-mantle !text-text !border-surface-0',
	}
}} />

{@render children()}
