import { Toast, toaster } from "@kobalte/core/toast";
import { X } from "lucide-solid";
import type { JSX } from "solid-js";

const baseClass =
	"flex items-start gap-3 rounded-lg border border-ctp-surface1 bg-ctp-surface0 p-3 text-sm text-ctp-text shadow-lg shadow-ctp-crust/50 data-opened:animate-toast-in data-closed:animate-toast-out";

const accentClass = {
	default: "border-l-4 border-l-ctp-sky",
	success: "border-l-4 border-l-ctp-green",
	error: "border-l-4 border-l-ctp-red",
} as const;

function body(message: string): JSX.Element {
	return (
		<>
			<span class="flex-1 leading-snug">{message}</span>
			<Toast.CloseButton
				class="rounded text-ctp-overlay1 transition-colors hover:cursor-pointer hover:text-ctp-text focus:outline-none focus:ring-2 focus:ring-ctp-sky/40"
				aria-label="Dismiss"
			>
				<X size={20} />
			</Toast.CloseButton>
		</>
	);
}

function show(message: string) {
	return toaster.show((props) => (
		<Toast
			toastId={props.toastId}
			class={`${baseClass} ${accentClass.default}`}
		>
			{body(message)}
		</Toast>
	));
}

function success(message: string) {
	return toaster.show((props) => (
		<Toast
			toastId={props.toastId}
			class={`${baseClass} ${accentClass.success}`}
		>
			{body(message)}
		</Toast>
	));
}

function error(message: string) {
	return toaster.show((props) => (
		<Toast toastId={props.toastId} class={`${baseClass} ${accentClass.error}`}>
			{body(message)}
		</Toast>
	));
}

function dismiss(id: number) {
	return toaster.dismiss(id);
}

export const toast = {
	show,
	success,
	error,
	dismiss,
};
