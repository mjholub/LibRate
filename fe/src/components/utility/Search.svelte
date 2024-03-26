<script lang="ts">
  import * as DropdownMenu from "$components/ui/dropdown-menu";
  import { Button } from "$components/ui/button";
	import axios from 'axios';
	import { _ } from 'svelte-i18n';
	import type { SearchResponse, resultCategory } from '$lib/types/search.ts';
	import type { CustomHttpError } from '$lib/types/error';
	import { SearchIcon } from 'svelte-feather-icons';
	import { goto } from '$app/navigation';
	import { Dropdown } from "@sveltestrap/sveltestrap";
	import Label from "$components/ui/label/label.svelte";
	import { Separator } from "bits-ui";
	import DropdownMenuCheckboxItem from "$components/ui/dropdown-menu/dropdown-menu-checkbox-item.svelte";
	import { page } from "$app/stores";
	import DropdownMenuSeparator from "$components/ui/dropdown-menu/dropdown-menu-separator.svelte";
	import DropdownMenuLabel from "$components/ui/dropdown-menu/dropdown-menu-label.svelte";
	import DropdownMenuRadioItem from "$components/ui/dropdown-menu/dropdown-menu-radio-item.svelte";
	import DropdownMenuRadioGroup from "$components/ui/dropdown-menu/dropdown-menu-radio-group.svelte";

	let query = '';
	let categories: string[] = [];
	let pageSize = 100;
	// TODO: allow modifying this once bleve search is fixed
	// let fuzzy = true;
	let categorySelections: Record<resultCategory, boolean> = {
		"artists": false,
    "genres": false,
    "media": false,
    "members": false,
    "ratings": false,
    "studios": false
	};
	let pageSizeString = "50";
	const pageSizeOptions = [10, 25, 50, 100];
	$: {
		categories = (Object.keys(categorySelections) as resultCategory[]).filter(category => categorySelections[category]);
		pageSize = parseInt(pageSizeString);
	}
	const categoriesList: resultCategory[] = ["artists", "genres", "media", "members", "ratings", "studios"]
	let errors: CustomHttpError[] = [];
	let showOptionsDropdown = false;

	async function searchItems(query: string) {
		try {
			const res = await axios.get<SearchResponse>('/api/search', {
				params: {
					q: query,
					category: categories.join(','),
					fuzzy: true,
					sort: 'Data.name',
					desc: true,
					page: 1,
					pageSize: pageSize
				},
			})
			if (res.status === 200) {
				goto('/search/results', { state: { results: res.data } });
			} else {
				errors.push({
					message: 'search request failed with status',
					status: res.status
				})
				errors = [...errors];
			}
		} catch (error) {
			errors.push({
					message: `search request failed with error: ${error}`,
					status: 500
			})
		}		
	}	
</script>

<div class="search-bar">
	<input
		type="text"
		class="search-input"
		bind:value={query}
		placeholder={$_('search_placeholder')}
		on:input={() => showOptionsDropdown = true}
		on:keydown={(e) => {
			if (e.key === 'Enter') {
				searchItems(query)
			}
		}}
	/>
		{#if showOptionsDropdown}
		<DropdownMenu.Root>
			<DropdownMenu.Trigger asChild let:builder>
				<Button variant="outline" builders={[builder]}>
					{$_('options')}
			</Button>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="w-56">
				<DropdownMenu.Label>
					{$_('categories')}
				</DropdownMenu.Label>	
				<DropdownMenu.Separator />
				{#each categoriesList as categoryItem}
					<DropdownMenuCheckboxItem bind:checked={categorySelections[categoryItem]}>
					{$_(categoryItem)}
					</DropdownMenuCheckboxItem>
				{/each}
				<DropdownMenuSeparator />
					<DropdownMenuLabel>
						{$_('page_size')}
					</DropdownMenuLabel>
					<DropdownMenuRadioGroup bind:value={pageSizeString}>
							{#each pageSizeOptions as pageSizeOpt}
						<DropdownMenuRadioItem value={pageSizeOpt.toString()}>
							{pageSizeOpt}
						</DropdownMenuRadioItem>
					{/each}
</DropdownMenuRadioGroup>
		</DropdownMenu.Content>
		</DropdownMenu.Root>
		{/if}
		<button on:click={() => searchItems(query)} id="search-button"><SearchIcon /></button>
</div>

<style>
	:root {
		--search-input-background: #ececec;
		--search-input-padding: 0.6em 1em;
		--search-box-outline: none;
	}

	.search-bar {
		width: 100%; /* initial width */
		padding-left: 2%;
		display: inline-flex;
	}

	button#search-button {
		display: inline-flex;
		align-items: center;
		position: sticky;
		width: 2em;
		border-radius: 0 4px 4px 0;
	}

	.search-input {
		background-color: var(--search-input-background);
		padding: var(--search-input-padding);
		border-radius: 4px 0 0 4px;
		outline: var(--search-box-outline);
	}

	@media (max-width: 768px) {
		.search-input {
			width: 70%;
			padding-left: 0.25em;
		}
	}

	@media (min-width: 768px) {
		.search-input:focus {
			transition: padding 1s ease-in-out;
			padding-right: 70%;
		}
	}

</style>
