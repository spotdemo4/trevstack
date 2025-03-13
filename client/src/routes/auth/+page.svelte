<script lang="ts">
	import { Tabs } from 'bits-ui';
</script>

<div class="pt-6">
	<Tabs.Root
		value="outbound"
		class="rounded-card border-muted bg-background-alt shadow-card w-[390px] border p-3"
	>
		<Tabs.List
			class="rounded-9px bg-dark-10 shadow-mini-inset dark:bg-background grid w-full grid-cols-2 gap-1 p-1 text-sm font-semibold leading-[0.01em] dark:border dark:border-neutral-600/30"
		>
			<Tabs.Trigger
				value="outbound"
				class="data-[state=active]:shadow-mini dark:data-[state=active]:bg-muted h-8 rounded-[7px] bg-transparent py-2 data-[state=active]:bg-white"
			>
				Outbound
			</Tabs.Trigger>
			<Tabs.Trigger
				value="inbound"
				class="data-[state=active]:shadow-mini dark:data-[state=active]:bg-muted h-8 rounded-[7px] bg-transparent py-2 data-[state=active]:bg-white"
			>
				Inbound
			</Tabs.Trigger>
		</Tabs.List>
		<Tabs.Content value="outbound" class="select-none pt-3">Test1</Tabs.Content>
		<Tabs.Content value="inbound" class="select-none pt-3">Test2</Tabs.Content>
	</Tabs.Root>
</div>
