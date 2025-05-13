<script lang="ts">
	import { buttonVariants } from '$lib/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { RangeCalendar as RangeCalendarPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: RangeCalendarPrimitive.DayProps = $props();

	export { className as class };
</script>

<RangeCalendarPrimitive.Day
	bind:ref
	class={cn(
		buttonVariants({ variant: 'ghost' }),
		'size-9 p-0 font-normal',
		'[&[data-today]:not([data-selected])]:bg-accent [&[data-today]:not([data-selected])]:text-crust [&[data-today]:not([data-selected])]:rounded-md',
		// Selected
		'data-[selected]:bg-surface data-[selected]:hover:bg-surface-1 data-[selected]:rounded-none data-[selected]:opacity-100',
		// Selection Start
		'data-[selection-start]:bg-surface-1 data-[selection-start]:text-text data-[selection-start]:hover:bg-surface-2 data-[selection-start]:rounded-l-md',
		// Selection End
		'data-[selection-end]:bg-surface-1 data-[selection-end]:text-text data-[selection-end]:hover:bg-surface-2 data-[selection-end]:rounded-r-md',
		// Outside months
		'data-[outside-month]:text-muted-foreground [&[data-outside-month][data-selected]]:bg-accent/50 [&[data-outside-month][data-selected]]:text-muted-foreground data-[outside-month]:pointer-events-none data-[outside-month]:opacity-50 [&[data-outside-month][data-selected]]:opacity-30',
		// Disabled
		'data-[disabled]:text-muted-foreground data-[disabled]:opacity-50',
		// Unavailable
		'data-[unavailable]:text-destructive-foreground data-[unavailable]:line-through',
		className
	)}
	{...restProps}
/>
