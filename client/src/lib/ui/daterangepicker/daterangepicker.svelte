<script lang="ts">
	import type { DateValue } from '@internationalized/date';
	import type { DateRange } from 'bits-ui';
	import { buttonVariants } from '$lib/ui/button';
	import * as Popover from '$lib/ui/popover';
	import { RangeCalendar } from '$lib/ui/range-calendar';
	import { CalendarDate, DateFormatter, getLocalTimeZone } from '@internationalized/date';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import { cn } from '$lib/utils.js';

	let {
		className,
		start,
		end,
		onchange
	}: {
		className?: string;
		start?: Date;
		end?: Date;
		onchange?: (start: Date, end: Date) => void;
	} = $props();

	let value: DateRange = $state({
		start: start
			? new CalendarDate(start.getFullYear(), start.getMonth(), start.getDate())
			: undefined,
		end: end ? new CalendarDate(end.getFullYear(), end.getMonth(), end.getDate()) : undefined
	});

	const df = new DateFormatter('en-US', {
		dateStyle: 'medium'
	});

	let startValue: DateValue | undefined = $state(undefined);
</script>

<div class={cn('grid gap-2', className)}>
	<Popover.Root>
		<Popover.Trigger
			class={cn(
				buttonVariants({ variant: 'input' }),
				'bg-based text-md md:text-sm',
				!value && 'text-subtext'
			)}
		>
			<CalendarIcon class="mr-2 size-4" />
			{#if value && value.start}
				{#if value.end}
					{df.format(value.start.toDate(getLocalTimeZone()))} - {df.format(
						value.end.toDate(getLocalTimeZone())
					)}
				{:else}
					{df.format(value.start.toDate(getLocalTimeZone()))}
				{/if}
			{:else if startValue}
				{df.format(startValue.toDate(getLocalTimeZone()))}
			{:else}
				Pick a date range
			{/if}
		</Popover.Trigger>
		<Popover.Content class="bg-based w-auto p-0" align="start">
			<RangeCalendar
				bind:value
				onStartValueChange={(v) => {
					startValue = v;
				}}
				onValueChange={(v) => {
					if (v.start && v.end) {
						onchange?.(v.start.toDate(getLocalTimeZone()), v.end.toDate(getLocalTimeZone()));
					}
				}}
				numberOfMonths={2}
			/>
		</Popover.Content>
	</Popover.Root>
</div>
