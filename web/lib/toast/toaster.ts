import { createToaster } from "@ark-ui/solid/toast";

export const toaster = createToaster({
	placement: "bottom-end",
	max: 5,
	gap: 12,
	overlap: false,
	offsets: "1.25rem",
});
