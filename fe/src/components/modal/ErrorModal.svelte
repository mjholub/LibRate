<script lang="ts">
	import { onMount } from 'svelte';
	import createError from 'http-errors';
	import type { CustomHttpError } from '$lib/types/error';
	export let showErrorModal = false;
	export let errorMessages: CustomHttpError[] = [];

	let dialog: HTMLDialogElement;

	// humanReadableCode converts a HTTP status code to a human readable string
	// e.g. 404 -> "Not Found"
	const humanReadableCode = (code: number): string => {
		return createError(code).message;
	};

	$: if (dialog) {
		dialog.setAttribute('aria-hidden', showErrorModal ? 'false' : 'true');

		if (showErrorModal) {
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
		showErrorModal = false;
	}}
	on:click|self={(e) => {
		if (e.target === dialog) {
			showErrorModal = false;
		}
	}}
	on:keydown|self={(e) => {
		if (e.key === 'Escape') {
			showErrorModal = false;
		}
	}}
>
	{#each errorMessages as errorMessage}
		<div on:click|stopPropagation role="dialog">
			<p class="error-string">
				{errorMessage.message}:
				<a href="https://http.cat/{errorMessage.status}">{humanReadableCode(errorMessage.status)}</a
				>
			</p>
		</div>
	{/each}
	<button on:click={() => (showErrorModal = false)}>OK</button>
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
		padding: 1em;
	}

	.error-string {
		display: flex;
		align-items: center;
		justify-content: center;
		padding-bottom: 0.5em;
		font-size: 7mm;
	}
</style>
