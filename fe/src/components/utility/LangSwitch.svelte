<script lang="ts">
	import { isLoading, locale, locales } from 'svelte-i18n';
	import { onMount } from 'svelte';

	const localeToName = new Map<string, string>([
		['en-US', 'English (U.S.)'],
		['pl', 'Polski'],
		['in', 'Indonesia'],
		['de', 'Deutsch']
	]);

	onMount(async () => {
		let count = 0;
		while (isLoading) {
			if (count > 25) {
				console.error('locales failed to load');
				break;
			}
			count++;
			await new Promise((resolve) => setTimeout(resolve, 100));
		}
		console.debug('locales loaded: ', locales);
		const storedLocale = localStorage.getItem('locale');
		if (storedLocale) {
			console.debug('setting locale to: ', storedLocale);
			locale.set(storedLocale);
		}
	});

	const handleLocaleChange = (e: Event) => {
		locale.set((e.target as HTMLSelectElement).value);
		localStorage.setItem('locale', (e.target as HTMLSelectElement).value);
	};
</script>

<div class="lang-switch">
	{#if $isLoading}
		<span class="dot-flashing" />
	{:else}
		<div class="dropdown-small">
			<select bind:value={$locale} on:change={handleLocaleChange}>
				{#each $locales as locale}
					{#if localeToName.has(locale)}
						<option value={locale}>{localeToName.get(locale)}</option>
					{:else}
						<option value={locale}>{locale}</option>
					{/if}
				{/each}
			</select>
		</div>
	{/if}
</div>

<style>
	:root {
		--dropdown-small-scale: 95%;
		--dot-flashing-color: #697ae0;
		--dot-flashing-fade: rgba(13, 29, 125, 0.3);
	}

	.dot-flashing {
		position: relative;
		width: 10px;
		height: 10px;
		border-radius: 5px;
		background-color: var(--dot-flashing-color);
		color: var(--dot-flashing-color);
		animation: dot-flashing 1s infinite linear alternate;
		animation-delay: 0.5s;
	}
	.dot-flashing::before,
	.dot-flashing::after {
		content: '';
		display: inline-block;
		position: absolute;
		top: 0;
	}
	.dot-flashing::before {
		left: -15px;
		width: 10px;
		height: 10px;
		border-radius: 5px;
		background-color: var(--dot-flashing-color);
		color: var(--dot-flashing-color);
		animation: dot-flashing 1s infinite alternate;
		animation-delay: 0s;
	}
	.dot-flashing::after {
		left: 15px;
		width: 10px;
		height: 10px;
		border-radius: 5px;
		background-color: var(--dot-flashing-color);
		color: var(--dot-flashing-color);
		animation: dot-flashing 1s infinite alternate;
		animation-delay: 1s;
	}

	@keyframes dot-flashing {
		0% {
			background-color: var(--dot-flashing-color);
		}
		50%,
		100% {
			background-color: var(--dot-flashing-fade);
		}
	}

	.lang-switch {
		position: relative;
		display: flex;
		float: right;
		scale: 95%;
	}
</style>
