<script lang="ts">
	import { onMount } from 'svelte';
	export let showSettingsModal = false;

	let dialog: HTMLDialogElement;

	$: if (dialog) {
		dialog.setAttribute('aria-hidden', showSettingsModal ? 'false' : 'true');

		if (showSettingsModal) {
			dialog.showModal();
		} else {
			dialog.close();
		}
	}

	onMount(() => {
		if (dialog) {
			dialog.setAttribute('aria-hidden', 'true');
			dialog.setAttribute('tabindex', '-1');
		}
	});
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
<dialog
	bind:this={dialog}
	aria-modal="true"
	on:close={() => {
		showSettingsModal = false;
	}}
	on:click|self={(e) => {
		if (e.target === dialog) {
			showSettingsModal = false;
		}
	}}
	on:keydown|self={(e) => {
		if (e.key === 'Escape') {
			showSettingsModal = false;
		}
	}}
>
	<div on:click|stopPropagation role="dialog">
		<slot name="settings" />
		<hr />
		<slot />
		<hr />
		<button on:click={() => (showSettingsModal = false)}>Close</button>
	</div>
</dialog>

<style>
	dialog {
		max-width: 32em;
		border-radius: 0.2em;
		border: none;
		padding: 0;
	}
	dialog::backdrop {
		background: rgba(0, 0, 0, 0.3);
	}
	dialog > div {
		padding: 1em;
	}
	button {
		display: block;
	}
</style>
