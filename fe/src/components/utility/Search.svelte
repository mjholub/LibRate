<script lang="ts">
	import { onMount } from 'svelte';
	import axios from 'axios';
	import type { SearchItem } from '$lib/types/search.ts';
	import { SearchIcon } from 'svelte-feather-icons';

	let search = '';
	let items: SearchItem[] | null = null;
	let error: string | null = null;

	async function searchItems() {
		// Check if the search field is empty
		if (search.trim() === '') {
			// If empty, clear the items and error
			items = null;
			error = null;
			return;
		}
		try {
			const response = await axios.post(
				'/api/search/',
				{ search },
				{
					headers: {
						'Content-Type': 'application/json'
					}
				}
			);

			items = response.data;
			error = null;
		} catch (err) {
			console.error('Error fetching data:', err);
			items = null;
			error = 'An error occurred while fetching data.';
		}
	}

	onMount(() => {
		searchItems();
	});
</script>

<div class="search-bar">
	<input
		type="text"
		class="search-input"
		bind:value={search}
		placeholder="Enter search keywords..."
		on:input={searchItems}
		on:keydown={(e) => {
			if (e.key === 'Enter') {
				searchItems();
			}
		}}
	/>
	<button on:click={searchItems} id="search-button"><SearchIcon /></button>
</div>

{#if error}
	<p>{error}</p>
{:else if items !== null}
	{#if items.length === 0}
		<p>No results found.</p>
	{:else}
		<div class="search-result-container">
			<p>Found {items.length} result{items.length === 1 ? '' : 's'}.</p>
			<ul class="search-results">
				{#each items as item (item.id)}
					<li class="search-result">{item.name}</li>
				{/each}
			</ul>
		</div>
	{/if}
{:else}
	<style>
		.search-result-container {
			display: none;
		}
	</style>
{/if}

<style>
	:root {
		--search-input-background: #ececec;
		--search-input-padding: 0.6em 1em;
	}

	.search-result-container {
		display: flex;
		flex-direction: column;
		align-items: start;
		padding-bottom: 0.5em;
	}

	.search-bar {
		margin-block-end: 0.2em;
		width: 100%; /* initial width */
		padding-left: 1em;
	}

	button#search-button {
		display: inline-block;
		position: absolute;
	}

	.search-input {
		background-color: var(--search-input-background);
		padding: var(--search-input-padding);
	}

	.search-input:focus {
		transition: padding 1s ease-in-out;
		padding-right: 70%;
	}

	.search-results {
		list-style-type: none;
		padding-left: 0;
		padding-bottom: 0.5em;
	}

	.search-result {
		margin-bottom: 0.5em;
	}
</style>
