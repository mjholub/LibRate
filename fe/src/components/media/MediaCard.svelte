<script lang="ts">
	import { onMount } from 'svelte';
	import type { Person } from '../../types/people.ts';
	import type { Media } from '../../types/media.ts';

	export const media: Media = {
		UUID: '',
		kind: '',
		name: '',
		genres: [],
		keywords: [],
		lang_ids: [],
		creators: []
	};
	export let title = '';
	export let image = '';
	export let averageRating = 0;
	export let creators = [] as Person[];

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
		{#each creators as creator (creator.id)}
			<div>{creator.first_name}</div>
			<div>{creator.last_name}</div>
		{/each}
	</div>
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
