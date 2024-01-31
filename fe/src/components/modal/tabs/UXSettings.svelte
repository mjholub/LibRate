<script defer lang="ts">
	import { _ } from 'svelte-i18n';
	import { fade } from 'svelte/transition';
	import { circOut } from 'svelte/easing';

	let themeSaved = false;
	let themes: string[] = ['default', 'light', 'sage', 'solarized', 'gruvbox', 'dracula'];
	let theme = localStorage.getItem('theme') || 'default';

	function setTheme() {
		themeSaved = true;
		localStorage.setItem('theme', theme);
	}
	function unsetTheme() {
		if (theme !== 'default') {
			localStorage.removeItem('theme');
		}
	}
</script>

<svelte:head>
	{#if theme === 'light'}
		<style>
			:root {
				--main-bg-color: #faeffe !important;
				--body-bgcolor: #faeffe !important;
				--text-color: #111 !important;
				--member-card-background-color: ghostwhite !important;
			}
		</style>
	{:else if theme === 'sage'}
		<style>
			:root {
				--main-bg-color: #487b63 !important;
				--body-bgcolor: #487b63 !important;
				--text-color: #202020 !important;
				--tertiary-text-color: #111515 !important;
				--member-card-background-color: #8b5848 !important;
				--member-card-color: #fgfefc !important;
				--button-bg: #94c1a6 !important;
				--button-radius: 16px !important;
				--icon-color: #d6f2c9 !important;
			}
		</style>
	{/if}
</svelte:head>

<label for="theme-selector">
	{$_('theme')}:
</label>
<select id="theme-selector" bind:value={theme}>
	{#each themes as theme}
		<option value={theme}>{theme}</option>
	{/each}
</select>
{#if themeSaved}
	<p transition:fade={{ delay: 500, duration: 500, easing: circOut }} class="confirmation-message">
		{$_('settings')}
		{$_('saved')}
	</p>
{/if}
<div class="save-reset-buttons">
	<button type="submit" class="save-button" on:click={setTheme}>
		{$_('save')}
	</button>
	<button type="submit" class="clear-button" on:click={unsetTheme}>
		{$_('reset')}
	</button>
</div>

<style>
	#theme-selector {
		margin-bottom: 1em;
	}

	button[type='submit'] {
		border-radius: var(--button-minor-border-radius);
	}
</style>
