import type { Virtualizer } from "@tanstack/solid-virtual";
import { LoaderCircle } from "lucide-solid";
import {
	type Component,
	createMemo,
	createResource,
	createSignal,
	Match,
	Show,
	Switch,
} from "solid-js";
import type { ListRequest, ListResponse } from "$connect/number/v1/list_pb";
import Splitter from "$lib/splitter";
import { NumberClient } from "$lib/transport";
import Form from "./Form";
import Table from "./Table";

type Request = Omit<ListRequest, "$typeName">;
type Response = Omit<ListResponse, "$typeName">;

const Numbers: Component = () => {
	const [request, setRequest] = createSignal<Request>({});

	const [response] = createResource<Response | undefined, Request>(
		request,
		async (req, info) => {
			const [resp, err] = await NumberClient.list(req);
			if (err) {
				console.error("Failed to fetch numbers:", err);
				return info.value;
			}

			if (resp.totalCount === info.value?.totalCount) {
				return {
					items: [...(info.value?.items ?? []), ...(resp.items ?? [])],
					totalCount: resp.totalCount,
					nextCursor: resp.nextCursor,
				};
			}

			return {
				items: resp.items,
				totalCount: resp.totalCount,
				nextCursor: resp.nextCursor,
			};
		},
	);

	const totalCount = createMemo(() => response()?.totalCount);

	const onScroll = (instance: Virtualizer<HTMLDivElement, Element>) => {
		const req = request();
		const resp = response.latest;

		if (!resp || !instance.range) return;
		if (resp.nextCursor === req?.cursor) return;
		if (instance.range.endIndex < resp.items.length - 5) return;

		setRequest({ ...req, cursor: resp.nextCursor });
	};

	return (
		<div class="h-body">
			<Splitter.Root
				panels={[
					{
						id: "a",
						minSize: 15,
						maxSize: 50,
					},
					{ id: "b" },
				]}
				defaultSize={[15, 50]}
			>
				<Splitter.Panel id="a" class="bg-ctp-mantle p-4">
					<Form onSubmit={(req) => setRequest(req)} />
				</Splitter.Panel>

				<Splitter.ResizeTrigger id="a:b" />

				<Splitter.Panel id="b">
					<Show
						when={totalCount()}
						fallback={
							<div class="flex h-full w-full items-center justify-center">
								<Switch>
									<Match when={totalCount() === undefined}>
										<LoaderCircle class="animate-spin" />
									</Match>
									<Match when={totalCount() === BigInt(0)}>
										<p>No numbers found :(</p>
									</Match>
								</Switch>
							</div>
						}
						keyed
					>
						{(resp) => (
							<Table
								count={resp}
								items={() => response()?.items ?? []}
								onScroll={onScroll}
							/>
						)}
					</Show>
				</Splitter.Panel>
			</Splitter.Root>
		</div>
	);
};

export default Numbers;
