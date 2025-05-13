<script lang="ts" module>
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAnchorAttributes, HTMLButtonAttributes } from 'svelte/elements';
	import type { VariantProps } from 'tailwind-variants';
	import { LoaderCircle } from '@lucide/svelte';
	import { cn } from '$lib/utils';
	import { tv } from 'tailwind-variants';

	export const buttonVariants = tv({
		base: cn(
			'inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-all',

			// Focus
			'focus-visible:outline-accent focus-visible:outline-2 focus-visible:outline-offset-2',

			// Disabled
			'disabled:pointer-events-none disabled:opacity-50',

			// Images
			"[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-5"
		),
		variants: {
			variant: {
				default: 'text-crust bg-accent hover:bg-accent/90 shadow-xs',
				red: 'text-crust bg-red hover:bg-red/90 shadow-xs',
				outline: 'text-text border-surface-1 hover:bg-surface border bg-transparent shadow-xs',
				input: 'text-text border-surface-1 hover:border-overlay border bg-transparent shadow-xs',
				ghost: 'text-text hover:bg-surface'
			},
			size: {
				default: 'h-9 px-4 py-2 has-[>svg]:px-3',
				sm: 'h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5',
				lg: 'h-10 rounded-md px-6 has-[>svg]:px-4',
				icon: 'h-9 min-w-9 px-3'
			}
		},
		defaultVariants: {
			variant: 'default',
			size: 'default'
		}
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
	export type ButtonSize = VariantProps<typeof buttonVariants>['size'];

	export type ButtonProps = WithElementRef<HTMLButtonAttributes> &
		WithElementRef<HTMLAnchorAttributes> & {
			variant?: ButtonVariant;
			size?: ButtonSize;
			loading?: boolean;
		};
</script>

<script lang="ts">
	let {
		class: className,
		variant = 'default',
		size = 'default',
		ref = $bindable(null),
		href = undefined,
		type = 'button',
		loading,
		children,
		...restProps
	}: ButtonProps = $props();
</script>

{#if href}
	<a
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{href}
		{...restProps}
	>
		{#if loading}
			<LoaderCircle class="animate-spin" />
		{/if}
		{@render children?.()}
	</a>
{:else}
	<button
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{type}
		{...restProps}
	>
		{#if loading}
			<LoaderCircle class="animate-spin" />
		{/if}
		{@render children?.()}
	</button>
{/if}
