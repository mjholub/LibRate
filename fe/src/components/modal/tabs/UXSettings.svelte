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
		<link rel="stylesheet" href="/static/themes/light.css" type="text/css" />
	{:else if theme === 'sage'}
		<link rel="stylesheet" href="/static/themes/sage.css" type="text/css" />
	{/if}</svelte:head
>

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
