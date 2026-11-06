export const modKey = navigator.userAgent.includes('Mac') ? '⌘' : 'Ctrl+';

export function isModKey(event: KeyboardEvent) {
	return event.metaKey || event.ctrlKey;
}

export function isEditableTarget(target: EventTarget | null) {
	if (!(target instanceof HTMLElement)) return false;
	return target.isContentEditable || !!target.closest('input, textarea, select');
}

export function hasOpenDialog() {
	return !!document.querySelector(
		'[data-slot="dialog-content"], [data-slot="alert-dialog-content"]'
	);
}
