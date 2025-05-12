<script>
	import '../app.css';
	import { Sonner } from '$lib/ui/sonner';
	import { ModeWatcher } from 'mode-watcher';

	let { children } = $props();
</script>

<ModeWatcher />
<Sonner />

{@render children()}
