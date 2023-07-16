<script lang="ts">
	import { onMount } from 'svelte';
	import { randomStore } from '../../stores/media/getRandom.ts';
	import { mediaImageStore } from '../../stores/media/image.ts';
	import { imageStore } from '../../stores/cdn/imagePath.ts';
	import MediaCard from './MediaCard.svelte';
	import AlbumCard from './AlbumCard.svelte';
	import type { Group, Person, Creator } from '../../types/people.ts';
	import type { AnyMedia, Media, MediaImage } from '../../types/media.ts';
	import type { Album, Track } from '../../types/music.ts';

	let media: (Album | Track | Media)[] = [];
	let album: Album = {
		UUID: '',
		kind: 'album',
		title: '',
		created: new Date(),
		media_id: '',
		name: '',
		album_artists: {
			person_artist: [],
			group_artist: []
		},
		creator: null,
		image_paths: [],
		release_date: new Date(),
		genres: [],
		keywords: [],
		duration: 0,
		tracks: []
	};
	let mediaImage: MediaImage = {
		mediaID: '',
		imageID: 0,
		isMain: false
	};
	let mediaImgPath = '';
	let creators: Creator[] = [];

	onMount(() => {
		(async () => {
			await randomStore.getRandom();
		})();

		console.info('mounting MediaCarousel initialized');

		let subscriptions: (() => void)[] = [];

		let unsubscribe = randomStore.subscribe((data) => {
			// FIXME: ehis is quite erroneouly expecting to retrieve sa simple media type,
			// which is not going to happen
			// instead, we need to refactor this code,
			// so that the values matching the media type, which the interfaces like Album
			// extend, are matched to the interface, and the rest is ignored or handled in a different way
			if (
				!data.mediaID ||
				!data.mediaTitle ||
				!data.mediaCreator ||
				!data.created ||
				!data.mediaKind
			) {
				console.warn('data is not valid: ', data);
				return;
			}

			const newMedia: Media = {
				UUID: data.mediaID[0],
				title: data.mediaTitle,
				kind: data.mediaKind,
				created: data.created,
				creator: data.mediaCreator
			};

			media = [...media, newMedia];
			console.debug('media: ', media);

			processMediaItems(media, subscriptions);
		});

		subscriptions.push(unsubscribe);

		return () => {
			subscriptions.forEach((unsub) => unsub());
		};
	});

	async function processMediaItems(mediaItems: (Media | Album)[], subscriptions: (() => void)[]) {
		for (const mediaItem of mediaItems) {
			console.debug('mediaItem: ', mediaItem);
			await mediaImageStore.getImagesByMedia(mediaItem.UUID);

			let mediaImgStrSub = mediaImageStore.subscribe((data) => {
				if (!data || !data.mediaID || data.images.length === 0) {
					return;
				}
				mediaImage = {
					mediaID: data.mediaID,
					imageID: data.images[0].imageID,
					isMain: data.mainImage.isMain
				};
			});

			subscriptions.push(mediaImgStrSub);

			await imageStore.getPaths(mediaImage.imageID);
			console.debug('imageStore: ', imageStore);

			let imgStoreSub = imageStore.subscribe((data) => {
				if (!data || !data.images || data.images.length === 0) {
					return;
				}
				mediaImgPath = data.images[0].source;
			});

			subscriptions.push(imgStoreSub);

			if (mediaItem.kind === 'album') {
				const album = mediaItem as Album;

				let creatorsArray = [
					...album.album_artists.person_artist,
					...album.album_artists.group_artist
				];

				for (const creator of creatorsArray) {
					if ('first_name' in creator) {
						const newPerson = creator as Person;
						const newCreator: Creator = { id: newPerson.id, name: newPerson.name };
						creators.push(newCreator);
					} else {
						const newGroup = creator as Group;
						const newCreator: Creator = { id: newGroup.id, name: newGroup.name };
						creators.push(newCreator);
					}
				}
			} else {
				const creatorArray = Array.isArray(mediaItem.creator)
					? mediaItem.creator
					: [mediaItem.creator];

				for (const creator of creatorArray) {
					if ('first_name' in creator) {
						const newPerson = creator as Person;
						const newCreator: Creator = { id: newPerson.id, name: newPerson.name };
						creators.push(newCreator);
					} else {
						const newGroup = creator as Group;
						const newCreator: Creator = { id: newGroup.id, name: newGroup.name };
						creators.push(newCreator);
					}
				}
			}
		}
	}

	const isAlbum = (mediaItem: Media | Album): mediaItem is Album => {
		return mediaItem.kind === 'album';
	};
	const isTrack = (mediaItem: Media | Track): mediaItem is Track => {
		return mediaItem.kind === 'track';
	};
</script>

<div class="carousel">
	{#if media.length === 0}
		<div>Loading...</div>
	{:else}
		{#each media as mediaItem (mediaItem.UUID)}
			<div class="media-card-wrapper">
				{#if isAlbum(mediaItem)}
					<AlbumCard {album} imgPath={mediaImgPath} />
				{:else if isTrack(mediaItem)}
					<p>Sorry, track cards are not yet implemented</p>
				{:else}
					<MediaCard media={mediaItem} title={mediaItem.title} image={mediaImgPath} {creators} />
				{/if}
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
