<script lang="ts">
	import axios from 'axios';
	import {
		Card,
		Label,
		Input,
		FormGroup,
		ListGroup,
		ListGroupItem
	} from '@sveltestrap/sveltestrap';
	// @ts-ignore
	import { getItem, setItem } from 'timedstorage';
	// @ts-ignore
	import * as time from 'timedstorage/time';
	import { PlusIcon } from 'svelte-feather-icons';
	import type { Album, AlbumArtist, Track } from '$lib/types/music';
	import { getMaxFileSize } from '$stores/form/upload';
	import { genreStore } from '$stores/media/genre';
	import { onMount, onDestroy } from 'svelte';
	import { openFilePicker } from '$stores/form/upload';
	import type { CustomHttpError } from '$lib/types/error';
	import type { NullableDuration } from '$lib/types/utils';
	// @ts-ignore
	import Tags from 'svelte-tags-input';
	import AddTrack from './AddTrack.svelte';

	export let nickname: string;
	let maxFileSize: number;
	let maxFileSizeString: string;
	let imagePaths: string[] = [];
	let errorMessages: CustomHttpError[] = [];
	let genreNames: string[] = [];
	let availableImportSources: string[] = [];
	let isUploading: boolean = false;
	let imageBase64 = '';
	let importSource = '';
	let importURL = '';

	let hasImportFinished = false;
	let remoteArtistsNames: string[] = [];
	let isArtistsListAmbiguous = false;
	let artistsToBeResolved: AlbumArtist[] = [];
	let album: Album = {
		UUID: '',
		kind: 'album',
		image_paths: [],
		media_id: '',
		name: '',
		title: '',
		created: new Date(),
		creator: null,
		creators: [],
		added: new Date(),
		album_artists: [],
		release_date: '',
		genres: [],
		duration: {
			Valid: false,
			Time: '00:00:00'
		},
		tracks: []
	};

	const shouldResetAlbum = () => {
		const albumWithLifeTime = localStorage.getItem('album');
		if (albumWithLifeTime) {
			const albumWithLifeTimeObject = JSON.parse(albumWithLifeTime);
			const albumLifeTime = albumWithLifeTimeObject.timeStamp;
			const currentTime = new Date().getTime();
			const timeDifference = currentTime - albumLifeTime;
			const timeDifferenceInHours = timeDifference / (1000 * 3600);
			if (timeDifferenceInHours > 1) {
				return true;
			}
		}
		return false;
	};

	onMount(async () => {
		try {
			if (getItem('genreNames')) {
				console.debug('getting genre names from local storage');
				genreNames = JSON.parse(getItem('genreNames') || '') || [];
				genreNames = [...genreNames];
				if (genreNames.length === 0) {
					console.debug('genre names array is empty, getting genre names from API endpoint');
					genreNames = await genreStore.getGenreNames('music', false);
					genreNames = [...genreNames];
					setItem('genreNames', JSON.stringify(genreNames), time.DAY);
				}
			} else {
				console.debug('getting genre names from API endpoint');
				genreNames = await genreStore.getGenreNames('music', false);
				genreNames = [...genreNames];
				setItem('genreNames', JSON.stringify(genreNames), time.DAY);
			}
		} catch (error) {
			console.error(error);
			errorMessages.push({
				message: 'Error getting genre links',
				status: 500
			});
		}
		const importSources = await fetch('/api/media/import-sources').then((res) => res.json());
		availableImportSources = importSources;
		availableImportSources = [...availableImportSources];
		maxFileSize = await getMaxFileSize();
		if (shouldResetAlbum()) {
			localStorage.removeItem('album');
		}
		addEventListener('load', () => {
			const dropArea = document.querySelector('.drop-area');
			if (dropArea) {
				dropArea.addEventListener('dragover', (e) => {
					e.preventDefault();
					dropArea.classList.add('highlight');
				});
				dropArea.addEventListener('dragleave', () => {
					dropArea.classList.remove('highlight');
				});
				dropArea.addEventListener('drop', async (e) => {
					e.preventDefault();
					dropArea.classList.remove('highlight');
					try {
						await updateImage(e);
					} catch (error) {
						errorMessages.push({
							message: "Couldn't upload image",
							status: 500
						});
						errorMessages = [...errorMessages];
					}
				});
			}
		});
		// save album, clear after 1 hour
		addEventListener('beforeunload', () => {
			let albumWithLifeTime: Object = {
				timeStamp: new Date().getTime(),
				data: album
			};
			localStorage.setItem('album', JSON.stringify(albumWithLifeTime));
		});
	});

	onDestroy(() => {
		maxFileSize = 0;
		isUploading = false;
		availableImportSources = [];
	});

	$: {
		maxFileSizeString = `${(maxFileSize / 1024 / 1024).toFixed(2)} MB`;
		imagePaths = album.image_paths || [];
		album.genres = album.genres || [];
		genreNames = [];
		isUploading = imagePaths.length !== 0;
	}

	const addMore = () => {};

	// updateImage reactively changes the displayed image when the user uploads a new one
	const updateImage = async (e: Event) => {
		const files = (e.target as HTMLInputElement).files;
		if (files) {
			const f = Array.from(files);
			f.forEach(async (file: File | Blob) => {
				if (file.size > maxFileSize) {
					errorMessages.push({
						message: `File size must be less than ${maxFileSizeString}`,
						status: 413
					});
					errorMessages = [...errorMessages];
				} else {
					imagePaths.push(URL.createObjectURL(file));
					imagePaths = [...imagePaths];
					await updateImageBase64(file);
				}
			});
		}
	};

	const updateImageBase64 = async (file: File | Blob): Promise<void> => {
		return new Promise((resolve, reject) => {
			const fileReader: FileReader = new FileReader();
			fileReader.onload = async () => {
				try {
					const imageBase64WithPrefix: string = fileReader.result as string;
					imageBase64 = imageBase64WithPrefix.split(',')[1]; // remove prefix
					isUploading = false;
					resolve();
				} catch (err) {
					reject(err);
				}
			};
			fileReader.onerror = (e) => reject(e);
			isUploading = true;
			fileReader.readAsDataURL(file);
		});
	};

	// submitImage sends the image to the server once the form has been filled out
	const submitImage = async (e: Event) => {
		return new Promise((resolve, reject) => {
			isUploading = true;
			console.debug('addImage');
			const files = (e.target as HTMLInputElement).files;
			if (files) {
				console.debug('files', files);
				const file = files[0];
				if (file.size > maxFileSize) {
					isUploading = false;
					reject(new Error(`File size must be less than ${maxFileSizeString}`));
				}
				const reader = new FileReader();
				console.info('file reader initialized');
				let csrfToken: string | undefined;
				csrfToken = document.cookie
					.split('; ')
					.find((row) => row.startsWith('csrf_'))
					?.split('=')[1];

				reader.onload = async (e) => {
					console.debug('reader onload');
					const data = e.target?.result;
					if (data) {
						console.debug('data', data);
						const formData = new FormData();
						console.info('form data initialized');
						formData.append('fileData', file);
						formData.append('imageType', 'album_cover');
						formData.append('member', nickname);
						try {
							const res = await axios.post('/api/upload/image', formData, {
								headers: {
									'Content-Type': 'multipart/form-data',
									Authorization: `Bearer ${localStorage.getItem('jwtToken')}`,
									'X-CSRF-Token': csrfToken || ''
								}
							});
							if (res.status === 200) {
								album.image_paths = [res.data.path];
								imagePaths = [res.data.path];
								isUploading = false;
								resolve(res.data.path);
							} else {
								errorMessages.push({
									message: res.data.message,
									status: res.status
								});
								isUploading = false;
								reject(res.status);
							}
						} catch (error) {
							errorMessages.push({
								message: 'Something went wrong',
								status: 500
							});
							isUploading = false;
							reject(error);
						}
					}
				};
				reader.readAsDataURL(file);
			}
		});
	};

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
	};

	const genreLinkFromName = (genreName: string): string => {
		const genreURI = genreName.toLowerCase().replace(' ', '-');
		const genreLink = window.location.origin + '/genres/music/' + genreURI;
		return genreLink;
	};

	const openGenreLink = (genreName: string) => window.open(genreLinkFromName(genreName), '_blank');

	const importFromWebSource = async (importSource: string, importURL: string) => {
		const csrfToken = document.cookie
			.split('; ')
			.find((row) => row.startsWith('csrf_'))
			?.split('=')[1];
		const res = await fetch('/api/media/import', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': csrfToken || ''
			},
			body: JSON.stringify({
				name: importSource,
				uri: importURL
			})
		});

		if (res.status !== 200) {
			errorMessages.push({
				message: 'Error importing album',
				status: res.status
			});
			console.error(errorMessages);
			errorMessages = [...errorMessages];
		}

		const albumData = await res.json();
		if ('album' in albumData && !('remote_artists' in albumData)) {
			album = albumData.album;
			artistsToBeResolved = albumData.artists;
			isArtistsListAmbiguous = true;
		} else if ('album' in albumData && 'remote_artists' in albumData) {
			remoteArtistsNames = albumData.remote_artists;
			album = albumData.album;
		} else {
			album = albumData;
		}
		hasImportFinished = true;
	};

	const importFromFile = async (e: Event) => {
		const files = (e.target as HTMLInputElement).files;
		if (files) {
			const f = Array.from(files);
			f.forEach(async (file: File | Blob) => {
				if (file.size > maxFileSize) {
					errorMessages.push({
						message: `File size must be less than ${maxFileSizeString}`,
						status: 413
					});
					errorMessages = [...errorMessages];
				} else {
					const fileReader: FileReader = new FileReader();
					fileReader.onload = async () => {
						// branch on mime type (application/json or audio/mpeg)
						switch (file.type) {
							case 'application/json':
								const json = fileReader.result as string;
								const albumJSON = JSON.parse(json);
								album.name = albumJSON.name;
								album.release_date = new Date(albumJSON.release_date);
								album.kind = 'album';
								// TODO: genre validation
								album.genres = albumJSON.genres;
								album.tracks = albumJSON.tracks;
								album.album_artists = albumJSON.album_artists;
								album.duration = sumAlbumDuration(albumJSON.tracks);
								break;
							case 'audio/mpeg':
							default:
								errorMessages.push({
									message: 'Invalid file type',
									status: 415
								});
								errorMessages = [...errorMessages];
								break;
						}
						fileReader.onerror = (e: ProgressEvent<FileReader>) => {
							errorMessages.push({
								message: 'Error reading file',
								status: 500
							});
							errorMessages = [...errorMessages];
							e.preventDefault();
						};
						fileReader.readAsText(file);
					};
				}
			});
		}
	};

	const sumAlbumDuration = (tracks: Track[]): NullableDuration => {
		let sum = 0;
		tracks.forEach((track) => {
			sum += track.duration as number;
		});
		return {
			Valid: true,
			// I hate this
			Time: sum as unknown as string
		};
	};

	const bufferSpotifyAlbumImage = async (imageUrl: string) => {
		const res = await fetch(imageUrl);
		const blob = await res.blob();
		const file = new File([blob], 'cover.jpg', { type: 'image/jpeg' });

		imagePaths.push(URL.createObjectURL(file));
		imagePaths = [...imagePaths];
		await updateImageBase64(file);
	};

	const enterImport = (e: KeyboardEvent) => {
		if (e.key === 'Enter') {
			e.preventDefault();
			importFromWebSource(importSource, importURL);
		} else {
			null;
		}
	};
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<div class="import-selector">
	<FormGroup>
		<Label id="import-label">Import from:</Label>
		<Input type="select" id="import-source" bind:value={importSource}>
			<option value="">Select a source</option>
			{#each availableImportSources as source}
				{#if source == 'rym'}
					<option value={source}>RateYourMusic</option>
				{:else if source == 'mediawiki'}
					<option value={source}>Wiki</option>
				{:else if source == 'id3'}
					<option value={source}>Music file (ID3 Tags)</option>
				{:else if source == 'json'}
					<option value={source}>JSON</option>
				{:else}
					<option value={source}>{source.charAt(0).toUpperCase() + source.slice(1)}</option>
				{/if}
			{/each}
		</Input>
		{#if importSource != '' && importSource != 'id3' && importSource != 'json'}
			{#if importSource == 'spotify'}
				<p aria-labelledby="spotify-info">Spotify album URL:</p>
			{:else if importSource == 'rym'}
				<p aria-labelledby="rym-info">RateYourMusic album URL:</p>
			{:else if importSource == 'discogs'}
				<p aria-labelledby="discogs-info">Discogs album URL:</p>
			{:else if importSource == 'lastfm'}
				<p aria-labelledby="lastfm-info">Last.fm album URL:</p>
			{:else if importSource == 'listenbrainz'}
				<p aria-labelledby="listenbrainz-info">ListenBrainz album URL:</p>
			{:else if importSource == 'bandcamp'}
				<p aria-labelledby="bandcamp-info">Bandcamp album URL:</p>
			{:else if importSource == 'mediawiki'}
				<p aria-labelledby="mediawiki-info">Wiki album URL:</p>
			{:else if importSource == 'pitchfork'}
				<p aria-labelledby="pitchfork-info">Pitchfork album URL:</p>
			{/if}
			<Input
				type="text"
				id="import-url"
				bind:value={importURL}
				on:keydown={enterImport}
				tabindex="0"
				role="input"
			/>
		{:else if importSource == 'json'}
			<p aria-labelledby="json-info">
				JSON (see <a href="https://codeberg.org/mjh/LibRate/wiki/Album-JSON-fields">specification</a
				>):
			</p>
			<Input type="file" id="import-file" on:change={importFromFile} />
		{:else if importSource == 'id3'}
			<p aria-labelledby="id3-info">Music file (ID3 Tags, only MP3 supported):</p>
			<Input type="file" id="import-file" on:change={importFromFile} />
		{/if}
	</FormGroup>
</div>

<!-- svelte-ignore  a11y-no-noninteractive-element-interactions -->
<div
	class="drop-area"
	on:drop={updateImage}
	on:click={() => openFilePicker(updateImage, 'image/*')}
	on:keydown={(e) => (e.key === 'Space' ? openFilePicker(updateImage, 'image/*') : null)}
	on:dragover={(e) => e.preventDefault()}
	aria-dropeffect="copy"
	role="region"
	aria-labelledby="drop-area-label"
>
	<p id="drop-area-label">
		<!-- svelte-ignore a11y-missing-attribute -->
		<a
			on:click={() => openFilePicker(updateImage, 'image/*')}
			on:keydown={(e) => (e.key === 'Enter' ? openFilePicker(updateImage, 'image/*') : null)}
			tabindex="0"
			role="button">Drop or click to add album cover here</a
		>
	</p>

	{#if isUploading}
		<div class="spinner" />
	{/if}
	{#if album.image_paths}
		<img src={album.image_paths[0]} alt="Album Cover" />
	{/if}
</div>

<label for="name">Album Name:</label>
<input id="name" bind:value={album.name} />

<label for="album-artists">Album Artists:</label>
{#if remoteArtistsNames.length > 0 && hasImportFinished}
	<p>The following artists were found in the import source, but not in the database:</p>
	<ListGroup>
		{#each remoteArtistsNames as artistName}
			<ListGroupItem>{artistName} [<a href="form/artist/add" target="_blank">Add</a>]</ListGroupItem
			>
		{/each}
	</ListGroup>
{/if}
{#if isArtistsListAmbiguous}
	<p>More than one artist was found for this album. Select one that matches.</p>
	<div class="artist-selection-grid">
		{#each artistsToBeResolved as artist}
			<Card class="artist-card">
				bind:value={album.album_artists}
				on:click={() => {
					album.album_artists = [artist];
					artistsToBeResolved = [];
					isArtistsListAmbiguous = false;
				}}
				on:keydown={(e) => {
					if (e.key === 'Enter') {
						album.album_artists = [artist];
						artistsToBeResolved = [];
						isArtistsListAmbiguous = false;
					}
				}}
				tabindex="0" role="button"
				<p><a href="/artists/${artist.artist}" target="_blank">{artist.name}</a></p>
			</Card>
		{/each}
	</div>
{/if}

<select id="album-artists" bind:value={album.album_artists} />

<button id="add-more" on:click={addMore}>
	<PlusIcon />
</button>

<label for="release-date">Release Date:</label>
<input id="release-date" bind:value={album.release_date} type="date" />

<label for="genres">Genres (comma separated):</label>
<div class="genre-box">
	<Tags
		bind:tags={album.genres}
		onlyUnique={true}
		autoComplete={localStorage.getItem('genreNames')
			? JSON.parse(localStorage.getItem('genreNames') || '')
			: []}
		onTagClick={openGenreLink}
		onlyAutocomplete={true}
	/>
</div>

<label for="duration">Duration:</label>
<input id="duration" bind:value={album.duration} type="time" />

<p>Tracks:</p>
<AddTrack />

<button on:click={handleSubmit}>Submit</button>

<style>
	.drop-area {
		border: 2px dashed #ccc;
		padding: 20px;
		text-align: center;
	}

	.drop-area img {
		max-width: 100%;
		max-height: 200px;
		margin-top: 10px;
	}

	.genre-box {
		display: inline-block;
		margin: 0 8px 8px 0;
		padding: 6px 12px;
		background-color: #f0f0f0;
		border: 1px solid #ccc;
		border-radius: 4px;
		position: relative;
	}

	.spinner {
		border: 4px solid rgba(0, 0, 0, 0.1);
		border-top: 4px solid #3498db;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		animation: spin 1s linear infinite;
		margin-left: 10px;
		display: inline-block;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>
