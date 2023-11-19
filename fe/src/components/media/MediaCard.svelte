<!-- NOTE: as is often the case with such generalised components,
this one might be completely scrapped in near future. -->
<script lang="ts">
	import { onMount } from 'svelte';
	import axios from 'axios';
	import type { Person, Group, Creator } from '$lib/types/people.ts';
	import type { Media } from '$lib/types/media.ts';

	export let media: Media = {
		UUID: '',
		kind: '',
		title: '',
		created: new Date(),
		creator: null,
		creators: [],
		added: new Date()
	};
	export let title = '';
	export let image = '';
	export let averageRating = 0;
	export const creators: Person[] | Group[] | Creator[] = [];

	// separate arrays for display purposes
	let individualCreators: Person[] = [];
	let groupCreators: Group[] = [];
	// processCreators checks the type of the creators prop
	function processCreators(creators: Person[] | Group[] | Creator[]) {
		if (creators.length === 0) {
			return;
		}

		// check if the first element is a Person
		if ('first_name' in creators[0]) {
			individualCreators = creators as Person[];
		} else {
			groupCreators = creators as Group[];
		}
	}

	const getAverageRatings = async () => {
		console.log('media.UUID', media.UUID);
		try {
			const response = await axios.get('/api/reviews/', {
				headers: {
					'Content-Type': 'application/json'
				},
				params: {
					id: JSON.stringify(media.UUID)
				}
			});

			averageRating = response.data.averageRating;
		} catch (error) {
			// Handle error here
			console.error('Error fetching average ratings:', error);
		}
	};

	onMount(() => {
		getAverageRatings();
		processCreators(creators);
	});
</script>

<div class="media-card">
	<img class="media-image" src={image} alt={title} />
	<div class="media-title">{title}</div>
	<div>Average rating: {averageRating}</div>
	<div class="media-creators">
		{#each individualCreators as person (person.id)}
			<div>{person.first_name}</div>
			<div>{person.last_name}</div>
		{/each}
		{#each groupCreators as group (group.id)}
			<div>{group.name}</div>
		{/each}
	</div>
	<!-- the following slot is used so that we can polymorphically render the media card 
with various types that extend from the Media iface-->
	<slot />
</div>

<style>
	.media-card {
		border: 1px solid #ccc;
		padding: 1em;
		margin: 1em;
	}

	.media-image {
		width: 100px;
	}

	.media-title {
		font-weight: bold;
		margin: 0.5em 0;
	}

	.media-creators {
		font-size: 0.9em;
		color: #666;
	}
</style>
