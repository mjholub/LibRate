<script lang="ts">
	import { _, locale } from 'svelte-i18n';
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import type { Member } from '$lib/types/member';
	import type { MemberPreferences } from '$stores/members/prefs';
	export let showSettingsModal = false;

	let dialog: HTMLDialogElement;
	export let member: Member;
	let themes: string[] = ['default', 'dark', 'nord', 'solarized', 'gruvbox', 'dracula'];
	// initialize with default values
	export let prefs: MemberPreferences = {
		locale: $locale,
		theme: 'default',
		auto_accept_follow: true,
		locally_searchable: true,
		robots_searchable: false,
		blur_nsfw: true,
		rating_scale_lower: 0,
		rating_scale_upper: 10,
		searchable_to_federated: true,
		message_autohide_words: [],
		muted_instances: []
	};

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
