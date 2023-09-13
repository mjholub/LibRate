<script lang="ts">
	import { onMount } from 'svelte';
	import axios from 'axios';
	import type { SearchItem } from '$lib/types/search.ts';

	let search = '';
	let items: SearchItem[] | null = null;
	let error: string | null = null;

	async function searchItems() {
		try {
			const response = await axios.post(
				'http://127.0.0.1:3000/api/search/',
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
		bind:value={search}
		placeholder="Enter search keywords..."
		on:input={searchItems}
		on:keydown={(e) => {
			if (e.key === 'Enter') {
				searchItems();
			}
		}}
	/>
	<button on:click={searchItems}>Search</button>
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
	.search-result-container {
		display: flex;
		flex-direction: column;
		align-items: left;
		padding-bottom: 0.5em;
	}

	.search-bar {
		margin-bottom: 0.8em;
		width: 50%; /* initial width */
		transition: width 0.4s ease-in-out;
	}

	.search-bar:focus {
		width: 80%;
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
