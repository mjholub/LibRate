<script lang="ts">
	import { onMount } from 'svelte';
	import { randomStore } from '../../stores/media/getRandom.ts';
	import MediaCard from './MediaCard.svelte';
	import type { Media } from '../../types/media.ts';

	let media: Media[] = [];

	onMount(async () => {
		await randomStore.getRandom();
		randomStore.subscribe((data) => {
			media = data.mediaID;
		});
	});
</script>

<div class="carousel">
	{#if media.length === 0}
		<div>Loading...</div>
	{:else}
		{#each media as mediaItem (mediaItem.UUID)}
			<div class="media-card-wrapper">
				<MediaCard
					title={mediaItem.name}
					image={mediaItem.image}
					averageRating={mediaItem.averageRating}
					creators={mediaItem.creators}
				/>
			</div>
		{/each}
	{/if}
</div>

<style>
	.carousel {
		display: flex;
		overflow-x: scroll;
		height: 100%;
		width: 100%;
	}

	.media-card-wrapper {
		flex: 0 0 auto;
		width: 30%;
		height: 40%;
		padding: 1em;
	}
</style>
