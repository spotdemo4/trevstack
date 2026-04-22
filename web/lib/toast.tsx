import { Toast, toaster } from "@kobalte/core/toast";

function show(message: string) {
	return toaster.show((props) => (
		<Toast toastId={props.toastId} class="toast">
			{message}
		</Toast>
	));
}

function success(message: string) {
	return toaster.show((props) => (
		<Toast toastId={props.toastId} class="toast toast--success">
			{message}
		</Toast>
	));
}

function error(message: string) {
	return toaster.show((props) => (
		<Toast toastId={props.toastId} class="toast toast--error">
			{message}
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
