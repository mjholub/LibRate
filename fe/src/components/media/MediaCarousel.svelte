<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { randomStore } from '$stores/media/getRandom.ts';
	import { mediaImageStore } from '$stores/media/image.ts';
	import { formatDuration } from '$stores/time/duration.ts';
	import AlbumCard from './AlbumCard.svelte';
	import FilmCard from './FilmCard.svelte';
	import type { MediaStoreState } from '$stores/media/media.ts';
	import type { Group, Person, Creator } from '$lib/types/people.ts';
	import type { Media } from '$lib/types/media.ts';
	import type { Album, Track } from '$lib/types/music.ts';
	import type { Book } from '$lib/types/books.ts';
	import type { Film } from '$lib/types/film_tv.ts';

	let media: (Album | Track | Film | Book)[] = [];
	let al: Album[] = [];
	let mediaImgPath = '';
	export let authenticated: boolean;
	let creators: Creator[] = [];

	const isAlbum = (mediaItem: Media | Album): mediaItem is Album => {
		return mediaItem.kind === 'album';
	};
	const isFilm = (mediaItem: Media | Film): mediaItem is Film => {
		return mediaItem.kind === 'film';
	};
	onMount(() => {
		initialFetch();
		console.info('mounting MediaCarousel initialized');

		subscribeToRandomStore();
	});

	onDestroy(() => {
		unsubscribeFromAll();
	});

	let subscriptions: (() => void)[] = [];

	async function initialFetch() {
		try {
			// first call to /api/media/random
			await randomStore.getRandom();
		} catch (error) {
			console.error('Error during initial fetch: ', error);
		}
	}

	function subscribeToRandomStore() {
		let isFirstUpdate = true;

		const unsubscribe = randomStore.subscribe((data: MediaStoreState) => {
			if (data.isLoading || !isFirstUpdate) {
				return;
			}
			if (!data.mediaType || data.media_id === null) {
				console.warn('data is not valid: ', data);
				return;
			}
			handleNewData(data);
			isFirstUpdate = false;
		});

		subscriptions.push(unsubscribe);
	}

	function handleNewData(data: MediaStoreState) {
		console.debug('handling new data: ', data);
		switch (data.mediaType) {
			case 'Album':
				const albumData = data.media as Album;
				let creator = null;
				if (albumData.album_artists && albumData.album_artists.person_artist.length > 0) {
					creator = albumData.album_artists.person_artist[0];
				} else if (albumData.album_artists && albumData.album_artists.group_artist.length > 0) {
					creator = albumData.album_artists.group_artist[0];
				}
				const albumCreators = albumData.creators ? albumData.creators : [];
				const addedDate = albumData.added ? albumData.added : new Date();
				const releaseDate = albumData.release_date
					? albumData.release_date.toString().split('T')[0]
					: null;
				const newAlbum: Album = {
					UUID: albumData.media_id,
					kind: 'album',
					title: albumData.name,
					created: new Date(),
					creator: creator,
					creators: albumCreators,
					added: addedDate,
					media_id: albumData.media_id,
					name: albumData.name,
					album_artists: albumData.album_artists || [],
					image_paths: albumData.image_paths,
					release_date: releaseDate,
					genres: albumData.genres,
					keywords: albumData.keywords,
					duration: albumData.duration,
					tracks: albumData.tracks || []
				};

				newAlbum.tracks = (albumData.tracks || []).map((track: Track) => {
					return {
						UUID: track.media_id,
						kind: 'track',
						title: track.name,
						created: new Date(),
						creator: null,
						creators: [],
						album_artists: creators || [],
						added: new Date(),
						media_id: track.media_id,
						name: track.name,
						album_id: track.album_id,
						duration: formatDuration(track.duration as string),
						lyrics: track.lyrics,
						track_number: track.track_number
					};
				});
				media.push(newAlbum);
				al.push(newAlbum);
				break;
			case 'Track':
				data.media = data.media as Track;
				const tracks = Array.isArray(data.media) ? data.media : [data.media];
				tracks.forEach((trackData) => {
					let newTrack: Track = {
						UUID: trackData.media_id,
						kind: 'track',
						title: trackData.name,
						created: new Date(),
						creator: null,
						creators: [],
						added: new Date(),
						media_id: trackData.media_id,
						name: trackData.name,
						album_id: trackData.album_id,
						duration: trackData.duration,
						lyrics: trackData.lyrics,
						track_number: trackData.track_number
					};

					media.push(newTrack);
				});
				break;
			case 'Book':
				data.media = data.media as Book;
				const books = Array.isArray(data.media) ? data.media : [data.media];
				const bookAddedDate = books[0].added ? books[0].added : new Date();
				books.forEach((bookData) => {
					let newBook: Book = {
						UUID: bookData.media_id,
						kind: 'book',
						title: bookData.title,
						created: bookData.publication_date,
						creator: bookData.authors[0],
						creators: bookData.authors,
						added: bookAddedDate,
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
				break;
			case 'Film':
				const filmData = data.media as Film;
				let newFilm: Film = {
					UUID: filmData.media_id,
					media_id: filmData.media_id,
					kind: 'film',
					title: filmData.title,
					created: new Date(),
					creator: null,
					creators: [],
					added: new Date(),
					castID: filmData.castID,
					synopsis: filmData.synopsis ? filmData.synopsis : 'No synopsis available.',
					releaseDate: filmData.releaseDate ? filmData.releaseDate : new Date(),
					duration: filmData.duration ? filmData.duration : null,
					rating: filmData.rating ? filmData.rating : 0
				};

				media.push(newFilm);
				break;
			case 'Unknown':
				console.warn('unknown media type: ', data.mediaType);
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

	async function processMediaItems(
		mediaItems: (Album | Film | Track | Book)[],
		subscriptions: (() => void)[]
	) {
		console.debug('staring mediaImageStore subscription');
		let mediaImgStrSub = mediaImageStore.subscribe((data) => {
			if (!data || data.mainImagePath === '') {
				return;
			}
			if (data.mainImagePath) {
				mediaImgPath = '.' + data.mainImagePath;
			}
		});

		subscriptions.push(mediaImgStrSub);
		console.debug('subscribed to mediaImageStore');
		for (const mediaItem of mediaItems) {
			await mediaImageStore.getImageByMedia(mediaItem.media_id);

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

			switch (mediaItem.kind) {
				case 'album':
					console.debug('mediaItem is an album');
					const album = mediaItem as Album;

					addCreators(album.album_artists.person_artist);
					addCreators(album.album_artists.group_artist);
				case 'film':
					console.debug('mediaItem is a film');
					const film = mediaItem as Film;

					if (film.castID) {
						// TODO: subscribe to cast store
					}
					break;
				case 'track':
					console.debug('mediaItem is a track. No creators to add.');
					break;
				case 'book':
					console.debug('mediaItem is a book');
					const book = mediaItem as Book;

					addCreators(book.authors);
					break;
				default:
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
	<div class="carousel-title">{$_('random_media')}:</div>
	<div class="carousel">
		{#if $randomStore.isLoading}
			<div>Loading...</div>
		{:else if media.length === 0}
			<div>No media items available.</div>
		{:else}
			{#each media as mediaItem (mediaItem.UUID)}
				<div class="media-card-wrapper">
					<p class="edit-prompt">
						{#if authenticated}
							<a href="/form/media/edit?${mediaItem.UUID}">(Edit this entry)</a>
						{:else}
							Log in to edit this entry.
						{/if}
					</p>

					{#if isAlbum(mediaItem)}
						{#each al as album (album)}
							<AlbumCard {album} imgPath={mediaImgPath} />
						{/each}
					{:else if isFilm(mediaItem)}
						<FilmCard posterPath={mediaImgPath} film={mediaItem} />
					{:else}
						<div class="hidden" />
					{/if}
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.edit-prompt {
		font-size: 0.8em;
		font-style: italic;
		margin-bottom: 0.5em;
		display: block;
	}

	.carousel-container {
		height: 100%;
		width: 100%;
	}
	.carousel-title {
		font-size: 1.5em;
		font-weight: bold;
		overflow: hidden;
	}
	.carousel {
		display: flex;
		position: relative;
		z-index: 2;
		overflow-x: scroll;
	}
	.media-card-wrapper {
		flex: 0 0 auto;
		width: 60%;
		height: 40%;
		padding: 1em;
	}
	.hidden {
		visibility: hidden;
	}
</style>
