<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { randomStore } from '../../stores/media/getRandom.ts';
	import { mediaImageStore } from '../../stores/media/image.ts';
	import { formatDuration } from '../../stores/time/duration.ts';
	import MediaCard from './MediaCard.svelte';
	import AlbumCard from './AlbumCard.svelte';
	import type { MediaStoreState } from '../../stores/media/media.ts';
	import type { Group, Person, Creator } from '../../types/people.ts';
	import type { Media } from '../../types/media.ts';
	import type { Album, Track } from '../../types/music.ts';
	import type { Book } from '../../types/books.ts';

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
	let al: Album[] = [];
	let mediaImgPath = '';
	let creators: Creator[] = [];

	const isAlbum = (mediaItem: Media | Album): mediaItem is Album => {
		return mediaItem.kind === 'album';
	};
	const isTrack = (mediaItem: Media | Track): mediaItem is Track => {
		return mediaItem.kind === 'track';
	};
	onMount(() => {
		initialFetch();
		console.info('mounting MediaCarousel initialized');

		subscribeToRandomStore();
	});

	onDestroy(() => {
		unsubscribeFromAll();
	});

	console.info('mounting MediaCarousel initialized');

	let subscriptions: (() => void)[] = [];

	async function initialFetch() {
		try {
			await randomStore.getRandom();
		} catch (error) {
			console.error('Error during initial fetch: ', error);
		}
	}

	function subscribeToRandomStore() {
		const unsubscribe = randomStore.subscribe((data: MediaStoreState) => {
			if (data.isLoading) {
				return;
			}
			if (!data.mediaType) {
				console.warn('data is not valid: ', data);
				return;
			}
			handleNewData(data);
		});

		subscriptions.push(unsubscribe);
	}

	function handleNewData(data: MediaStoreState) {
		switch (data.mediaType) {
			case 'Album':
				if (data.album) {
					const albums = Array.isArray(data.album) ? data.album : [data.album];
					albums.forEach((albumData) => {
						let creator = null;
						if (albumData.album_artists && albumData.album_artists.person_artist.length > 0) {
							creator = albumData.album_artists.person_artist[0];
						} else if (albumData.album_artists && albumData.album_artists.group_artist.length > 0) {
							creator = albumData.album_artists.group_artist[0];
						}
						const releaseDate = albumData.release_date
							? albumData.release_date.toString().split('T')[0]
							: null;
						let newAlbum: Album = {
							UUID: albumData.media_id,
							kind: 'album',
							title: albumData.name,
							created: new Date(),
							creator: creator,
							media_id: albumData.media_id,
							name: albumData.name,
							album_artists: albumData.album_artists,
							image_paths: albumData.image_paths,
							release_date: releaseDate,
							genres: albumData.genres,
							keywords: albumData.keywords,
							duration: albumData.duration,
							tracks: albumData.tracks
						};

						newAlbum.tracks.forEach((track) => {
							track.duration = formatDuration(track.duration as string);
						});
						media.push(newAlbum);
						al.push(newAlbum);
					});
				}
				break;
			case 'Track':
				if (data.track) {
					const tracks = Array.isArray(data.track) ? data.track : [data.track];
					tracks.forEach((trackData) => {
						let newTrack: Track = {
							UUID: trackData.media_id,
							kind: 'track',
							title: trackData.name,
							created: new Date(),
							creator: null,
							media_id: trackData.media_id,
							name: trackData.name,
							album_id: trackData.album_id,
							duration: trackData.duration,
							lyrics: trackData.lyrics,
							track_number: trackData.track_number
						};

						media.push(newTrack);
					});
				}
				break;
			case 'Book':
				if (data.book) {
					const books = Array.isArray(data.book) ? data.book : [data.book];
					books.forEach((bookData) => {
						let newBook: Book = {
							UUID: bookData.media_id,
							kind: 'book',
							title: bookData.title,
							created: bookData.publication_date,
							creator: bookData.authors[0],
							authors: bookData.authors,
							media_id: bookData.media_id,
							publisher: bookData.publisher,
							publication_date: bookData.publication_date,
							genres: bookData.genres,
							keywords: bookData.keywords,
							isbn: bookData.isbn,
							asin: bookData.asin,
							pages: bookData.pages,
							cover: bookData.cover,
							summary: bookData.summary,
							languages: bookData.languages
						};

						media.push(newBook);
					});
				}
				break;
			case 'Film':
				break;
			case 'Unknown':
				break;
			default:
				console.warn('unknown media type: ', data.mediaType);
				break;
		}

		console.debug('media: ', media);

		processMediaItems(media, subscriptions);
	}

	function unsubscribeFromAll() {
		subscriptions.forEach((unsub) => unsub());
		subscriptions = [];
	}

	async function processMediaItems(mediaItems: (Media | Album)[], subscriptions: (() => void)[]) {
		for (const mediaItem of mediaItems) {
			console.debug('mediaItem: ', mediaItem);
			await mediaImageStore.getImageByMedia(mediaItem.UUID);

			console.debug('staring mediaImageStore subscription');
			let mediaImgStrSub = mediaImageStore.subscribe((data) => {
				if (!data || data.mainImagePath === '') {
					return;
				}
				if (data.mainImagePath) {
					mediaImgPath = '.' + data.mainImagePath;
					console.debug('mediaImgPath: ', mediaImgPath);
				}
			});

			subscriptions.push(mediaImgStrSub);
			console.debug('subscribed to mediaImageStore');

			/*await imageStore.getPaths(mediaImage.imageID);
			console.debug('image paths for media ID: ', mediaImage.imageID, mediaImage.mediaID);

			console.debug('staring imageStore subscription');
			let imgStoreSub = imageStore.subscribe((data) => {
				if (!data || !data.images || data.images.length === 0) {
					return;
				}
				mediaImgPath = data.images[0].source;
			});

			subscriptions.push(imgStoreSub);
			console.debug('subscribed to imageStore');

        */
			const addCreators = (creatorArray: (Person | Group)[]) => {
				for (const creator of creatorArray) {
					let newCreator: Creator;
					if ('first_name' in creator) {
						const newPerson = creator as Person;
						newCreator = {
							id: newPerson.id,
							name: newPerson.first_name + ' ' + newPerson.last_name
						};
					} else {
						const newGroup = creator as Group;
						newCreator = { id: newGroup.id, name: newGroup.name };
					}
					creators.push(newCreator);
					console.debug('added creator: ', newCreator);
				}
			};

			if (isAlbum(mediaItem)) {
				console.debug('mediaItem is an album');
				const album = mediaItem as Album;

				addCreators(album.album_artists.person_artist);
				addCreators(album.album_artists.group_artist);
			} else {
				console.debug('mediaItem is not an album');
				const media = mediaItem as Media;

				if (media.creator) {
					creators.push(media.creator);
				}
			}
		}
	}
</script>

<div class="carousel-container">
	<div class="carousel-title">Randomly selected media:</div>
	<div class="carousel">
		{#if $randomStore.isLoading}
			<div>Loading...</div>
		{:else if media.length === 0}
			<div>No media items available.</div>
		{:else}
			{#each media as mediaItem (mediaItem.UUID)}
				<div class="media-card-wrapper">
					{#if isAlbum(mediaItem)}
						{#each al as album (album)}
							<AlbumCard {album} imgPath={mediaImgPath} />
						{/each}
					{:else if isTrack(mediaItem)}
						<p>Sorry, track cards are not yet implemented</p>
					{:else}
						<MediaCard media={mediaItem} title={mediaItem.title} image={mediaImgPath} {creators} />
					{/if}
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.carousel-container {
		height: 100%;
		width: 100%;
	}
	.carousel-title {
		font-size: 1.5em;
		font-weight: bold;
		margin-bottom: 1em;
		overflow: hidden;
	}
	.carousel {
		display: flex;
		overflow-x: scroll;
	}
	.media-card-wrapper {
		flex: 0 0 auto;
		width: 30%;
		height: 40%;
		padding: 1em;
	}
</style>
