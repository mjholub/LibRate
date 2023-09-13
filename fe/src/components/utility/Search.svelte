<script lang="ts">
	import { onMount } from 'svelte';
	import type { SearchItem } from '$lib/types/search.ts';

	let search = '';
	let items: SearchItem[] = [];

	// function to fetch data from the backend based on the search term
	async function searchItems() {
		const response = await fetch('http://localhost:3000/api/search', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ search })
		});

		const data = await response.json();

		// If the backend responds with the filtered items
		// update items with the new data
		items = data;
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

<div>
	<ul class="search-results">
		{#each items as item (item.id)}
			<li class="search-result">{item.name}</li>
		{/each}
	</ul>
</div>

<style>
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
	}

	.search-result {
		margin-bottom: 0.5em;
	}
</style>
