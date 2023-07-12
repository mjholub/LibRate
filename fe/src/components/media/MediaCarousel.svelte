<script lang="ts">
	import { onMount } from 'svelte';
	import { randomStore } from '../../stores/media/getRandom.ts';
	import { mediaImageStore } from '../../stores/media/media.ts';
	import { imageStore } from '../../stores/cdn/imagePath.ts';
	import MediaCard from './MediaCard.svelte';
	import type { Person } from '../../types/people.ts';
	import type { Media, MediaImage } from '../../types/media.ts';

	let media: Media[] = [];
	let mediaImage: MediaImage = {
		mediaID: '',
		imageID: 0,
		isMain: false
	};
	let mediaImgPath = '';
	let creators = [] as Person[];

	onMount(async () => {
		await randomStore.getRandom();
		randomStore.subscribe((data) => {
			if (
				!data.mediaID ||
				!data.mediaTitle ||
				!data.mediaCreator ||
				!data.created ||
				!data.mediaKind
			) {
				return;
			}
			const newMedia: Media = {
				UUID: data.mediaID[0],
				title: data.mediaTitle,
				kind: data.mediaKind,
				created: data.created,
				creator: data.mediaCreator
			};
			media.push(newMedia);
		});
		media.forEach(async (mediaItem) => {
			await mediaImageStore.getImagesByMedia(mediaItem.UUID);
			mediaImageStore.subscribe((data) => {
				if (!data || !data.mediaID || data.images.length === 0) {
					return;
				}
				mediaImage = {
					mediaID: data.mediaID,
					imageID: data.images[0].imageID,
					isMain: data.mainImage.isMain
				};
			});
			// fetch the paths of the images using getPaths from imageStore
			await imageStore.getPaths(mediaImage.imageID);
			imageStore.subscribe((data) => {
				if (!data || !data.images || data.images.length === 0) {
					return;
				}
				mediaImgPath = data.images[0].source;
			});
			if (mediaItem?.creator) {
				creators.push(mediaItem?.creator);
			}
		});
	});
</script>

<div class="carousel">
	{#if media.length === 0}
		<div>Loading...</div>
	{:else}
		{#each media as mediaItem (mediaItem.UUID)}
			<div class="media-card-wrapper">
				<MediaCard title={mediaItem.title} image={mediaImgPath} {creators} />
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
