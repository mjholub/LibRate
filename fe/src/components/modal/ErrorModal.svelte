<script lang="ts">
	import { onMount } from 'svelte';
	import type { CustomHttpError } from '$lib/types/error';
	export let showErrorModal = false;
	export let errorMessages: CustomHttpError[] = [];
	let currentIndex = 0;

	let dialog: HTMLDialogElement;

	const nextError = () => {
		currentIndex = (currentIndex + 1) % errorMessages.length;
	};

	const prevError = () => {
		currentIndex = (currentIndex - 1 + errorMessages.length) % errorMessages.length;
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
	{#each errorMessages as errorMessage, index}
		{#if errorMessages.length > 1}
			<div class="error-navigator">
				<button on:click={prevError} disabled={currentIndex === 0}>Previous</button>
				<span>Error {currentIndex + 1}/{errorMessages.length}</span>
				<button on:click={nextError} disabled={currentIndex === errorMessages.length - 1}
					>Next</button
				>
			</div>
		{/if}
		<div on:click|stopPropagation role="dialog">
			<p class="error-string">
				<a href={`https://http.cat/{errorMessage.status}`}>
					<img
						id="httpcat"
						src={`https://http.cat/${errorMessage.status}`}
						alt={`HTTP Status ${errorMessage.status}`}
						crossorigin="anonymous"
					/>
				</a>
			</p>
			<p class="error-details">{errorMessage.message}</p>
		</div>
	{/each}
	<button on:click={() => (showErrorModal = false)}>OK</button>
</dialog>

<style>
	#httpcat {
		margin: 10%;
		display: block;
		position: sticky;
		max-width: 95%;
	}

	dialog {
		max-width: calc(32em + 10%);
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

	.error-navigator {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5em;
	}
</style>
