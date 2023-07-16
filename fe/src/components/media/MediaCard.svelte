<script lang="ts">
	import { onMount } from 'svelte';
	import type { Either } from 'typescript-monads';
	import type { Person, Group } from '../../types/people.ts';
	import type { Media } from '../../types/media.ts';

	export let media: Media = {
		UUID: '',
		kind: '',
		title: '',
		created: new Date(),
		creator: null
	};
	export let title = '';
	export let image = '';
	export let averageRating = 0;
	// implement polymorphism by using Either type
	// this is needed to match the fields from e.g. the album type
	// when e.g. the album card renders this component
	export let creators: Either<Person[], Group[]>;
	// separate arrays for display purposes
	let individualCreators: Person[] = [];
	let groupCreators: Group[] = [];

	creators.match<Person[] | Group[]>({
		left: (people: Person[]) => (individualCreators = people),
		right: (groups: Group[]) => (groupCreators = groups)
	});

	const getAverageRatings = async () => {
		const response = await fetch('/api/media/averageRatings', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ mediaId: media.UUID })
		});
		const data = await response.json();
		averageRating = data.averageRating;
	};

	onMount(() => {
		getAverageRatings();
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
