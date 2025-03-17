<script lang="ts">
	import { ArrowLeft, ArrowRight, Minus, Calendar, X } from '@lucide/svelte';
	import { DateRangePicker, type DateRange } from 'bits-ui';
	import { fade } from 'svelte/transition';
	import { getLocalTimeZone } from '@internationalized/date';
	import { cn } from '$lib/utils';

	let {
		className,
		start = $bindable(),
		end = $bindable(),
		onchange
	}: {
		className?: string;
		start?: Date;
		end?: Date;
		onchange?: (start?: Date, end?: Date) => void;
	} = $props();

	let daterange: DateRange = $state({
		start: undefined,
		end: undefined
	});
	let rerender = $state(false);
</script>

<!-- Need to rerender because setting to undefined doesn't work  -->
{#key rerender}
	<DateRangePicker.Root
		bind:value={daterange}
		onValueChange={(v) => {
			if (v.start && v.end) {
				start = v.start.toDate(getLocalTimeZone());
				end = v.end.toDate(getLocalTimeZone());
				if (onchange) {
					onchange(start, end);
				}
			}
		}}
		class={cn(className)}
	>
		<div
			class="bg-mantle border-surface-0 hover:border-surface-2 flex items-center rounded border pl-2 text-sm drop-shadow-md transition-all"
		>
			<div class="grow flex items-center justify-center">
				{#each ['start', 'end'] as const as type}
					<DateRangePicker.Input {type}>
						{#snippet children({ segments })}
							{#each segments as { part, value }}
								<div class="inline-block select-none">
									{#if part === 'literal'}
										<DateRangePicker.Segment {part} class="text-overlay-0 p-1">
											{value}
										</DateRangePicker.Segment>
									{:else}
										<DateRangePicker.Segment
											{part}
											class="aria-[valuetext=Empty]:text-overlay-0 hover:bg-surface-0 focus:bg-surface-0 focus:outline-sky rounded p-0.5 transition-all focus:outline focus:outline-offset-1"
										>
											{value}
										</DateRangePicker.Segment>
									{/if}
								</div>
							{/each}
						{/snippet}
					</DateRangePicker.Input>
					{#if type === 'start'}
						<div aria-hidden="true" class="px-1">
							<Minus size="10" />
						</div>
					{/if}
				{/each}
			</div>
			<DateRangePicker.Trigger
				class="text-overlay-2 hover:bg-surface-0 grow flex justify-center items-center focus:outline-sky ml-1 cursor-pointer p-2 transition-all focus:outline focus:outline-offset-1"
			>
				<Calendar size="20" />
			</DateRangePicker.Trigger>
			<button
				class="text-overlay-2 hover:bg-surface-0 focus:outline-sky cursor-pointer rounded-r p-2 transition-all focus:outline focus:outline-offset-1"
				onclick={() => {
					if (daterange) {
						daterange.end = undefined;
						daterange.start = undefined;
					}
					start = undefined;
					end = undefined;
					if (onchange) {
						onchange(start, end);
					}
					rerender = !rerender;
				}}
			>
				<X size="20" />
			</button>
		</div>
		<DateRangePicker.Content forceMount>
			{#snippet child({ props, open })}
				{#if open}
					<div
						{...props}
						class="absolute z-50"
						transition:fade={{
							duration: 100
						}}
					>
						<DateRangePicker.Calendar
							class="border-surface-0 bg-mantle mt-1 rounded border p-3 drop-shadow-md"
						>
							{#snippet children({ months, weekdays })}
								<DateRangePicker.Header class="flex items-center justify-between">
									<DateRangePicker.PrevButton
										class="hover:bg-surface-0 inline-flex size-10 cursor-pointer items-center justify-center rounded transition-all active:scale-[0.98]"
									>
										<ArrowLeft />
									</DateRangePicker.PrevButton>
									<DateRangePicker.Heading class="select-none font-medium" />
									<DateRangePicker.NextButton
										class="hover:bg-surface-0 inline-flex size-10 cursor-pointer items-center justify-center rounded transition-all active:scale-[0.98]"
									>
										<ArrowRight />
									</DateRangePicker.NextButton>
								</DateRangePicker.Header>
								<div class="flex flex-col space-y-4 pt-4 sm:flex-row sm:space-x-4 sm:space-y-0">
									{#each months as month}
										<DateRangePicker.Grid class="w-full border-collapse select-none space-y-1">
											<DateRangePicker.GridHead>
												<DateRangePicker.GridRow class="mb-1 flex w-full justify-between">
													{#each weekdays as day}
														<DateRangePicker.HeadCell
															class="text-overlay-0 font-normal! w-10 rounded text-xs"
														>
															{day.slice(0, 2)}
														</DateRangePicker.HeadCell>
													{/each}
												</DateRangePicker.GridRow>
											</DateRangePicker.GridHead>
											<DateRangePicker.GridBody>
												{#each month.weeks as weekDates}
													<DateRangePicker.GridRow class="flex w-full">
														{#each weekDates as date}
															<DateRangePicker.Cell
																{date}
																month={month.value}
																class="p-0! relative m-0 size-10 overflow-visible text-center text-sm focus-within:relative focus-within:z-20"
															>
																<DateRangePicker.Day
																	class={'hover:border-sky focus-visible:ring-foreground! data-selected:rounded-none data-selection-end:rounded-r data-selection-start:rounded-l data-highlighted:bg-surface-0 data-selected:bg-surface-1 data-selection-end:bg-surface-2 data-selection-start:bg-surface-2 data-disabled:text-text/30 data-unavailable:text-overlay-0 data-disabled:pointer-events-none data-outside-month:pointer-events-none data-highlighted:rounded-none data-unavailable:line-through group relative inline-flex size-10 items-center justify-center overflow-visible whitespace-nowrap rounded border border-transparent bg-transparent p-0 text-sm font-normal transition-all'}
																>
																	<div
																		class="bg-sky group-data-selected:bg-background group-data-today:block absolute top-[5px] hidden size-1 rounded-full transition-all"
																	></div>
																	{date.day}
																</DateRangePicker.Day>
															</DateRangePicker.Cell>
														{/each}
													</DateRangePicker.GridRow>
												{/each}
											</DateRangePicker.GridBody>
										</DateRangePicker.Grid>
									{/each}
								</div>
							{/snippet}
						</DateRangePicker.Calendar>
					</div>
				{/if}
			{/snippet}
		</DateRangePicker.Content>
	</DateRangePicker.Root>
{/key}
